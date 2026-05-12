package evaluator

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"scql/ast"
	"strings"
)

// Evaluator evaluates the AST and performs the actual data fetching.
type Evaluator struct{}

func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate evaluates a program and returns a list of results (one per statement).
// For now we will focus on SELECT statements.
func (e *Evaluator) Evaluate(program *ast.Program) ([]interface{}, error) {
	var results []interface{}

	for _, stmt := range program.Statements {
		res, err := e.evaluateStatement(stmt)
		if err != nil {
			return nil, err
		}
		if res != nil {
			results = append(results, res)
		}
	}

	return results, nil
}

func (e *Evaluator) evaluateStatement(stmt ast.Statement) (interface{}, error) {
	switch s := stmt.(type) {
	case *ast.SelectStatement:
		return e.evaluateSelectStatement(s)
		// Additional statements like SET, AUTHENTICATE would be implemented here
	}
	return nil, nil
}

func stripQuotes(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[0] == s[len(s)-1] {
		return s[1 : len(s)-1]
	}
	return s
}

func (e *Evaluator) evaluateSelectStatement(s *ast.SelectStatement) (interface{}, error) {
	// 1. Get Target URL
	var targetURL string
	if fromStr, ok := s.From.(*ast.StringLiteral); ok {
		targetURL = stripQuotes(fromStr.Value)
	} else if fromIdent, ok := s.From.(*ast.Identifier); ok {
		targetURL = stripQuotes(fromIdent.Value)
	} else {
		return nil, fmt.Errorf("invalid FROM clause: must be a URL string")
	}

	fetchURL := targetURL
	// Make HTTP GET request
	if !strings.HasPrefix(fetchURL, "http://") && !strings.HasPrefix(fetchURL, "https://") {
		// Default to https if missing
		fetchURL = "https://" + fetchURL
	}

	resp, err := http.Get(fetchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	// 2. Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// 3. Determine Field Selectors and Aliases
	type fieldMeta struct {
		Name     string
		Selector string
	}

	fields := make([]fieldMeta, 0)

	for _, field := range s.Fields {
		var name string
		var selector string

		switch f := field.(type) {
		case *ast.Identifier:
			// Regular column, e.g. SELECT ".title"
			name = f.Value
			selector = f.Value

		case *ast.StringLiteral:
			// e.g. SELECT ".title"
			name = stripQuotes(f.Value)
			selector = stripQuotes(f.Value)

		case *ast.AliasExpression:
			// Alias, e.g. SELECT ".title" AS title
			name = f.Alias
			if leftStr, ok := f.Left.(*ast.StringLiteral); ok {
				selector = stripQuotes(leftStr.Value)
			} else if leftIdent, ok := f.Left.(*ast.Identifier); ok {
				selector = leftIdent.Value
			} else {
				return nil, fmt.Errorf("alias expression left side must be a string selector")
			}
		default:
			return nil, fmt.Errorf("unsupported field expression: %s", field.String())
		}

		fields = append(fields, fieldMeta{Name: name, Selector: selector})
	}

	// 4. Extract Data
	// For simplicity, we assume we find multiple elements for each selector and group them into rows.
	// A better way would be finding a "container" element first, but for now we find all elements.
	// Actually, wait, it's better to get the maximum number of items found for any field.

	columnData := make(map[string][]string)
	maxLen := 0

	for _, f := range fields {
		var items []string
		doc.Find(f.Selector).Each(func(i int, s *goquery.Selection) {
			items = append(items, strings.TrimSpace(s.Text()))
		})
		columnData[f.Name] = items
		if len(items) > maxLen {
			maxLen = len(items)
		}
	}

	// Zip columns into rows
	var rows []map[string]interface{}
	for i := 0; i < maxLen; i++ {
		row := make(map[string]interface{})
		for _, f := range fields {
			if i < len(columnData[f.Name]) {
				row[f.Name] = columnData[f.Name][i]
			} else {
				row[f.Name] = nil
			}
		}
		rows = append(rows, row)
	}

	// 5. WHERE clause (Simplified for now)
	// We can implement actual AST evaluation for WHERE later.

	return rows, nil
}
