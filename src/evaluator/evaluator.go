package evaluator

import (
	"fmt"
	"net/http"
	"regexp"
	"scql/ast"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
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
	var rows []map[string]interface{}

	if s.Rows != nil {
		// Scoped extraction
		var rowsSelector string
		if rowStr, ok := s.Rows.(*ast.StringLiteral); ok {
			rowsSelector = stripQuotes(rowStr.Value)
		} else if rowIdent, ok := s.Rows.(*ast.Identifier); ok {
			rowsSelector = rowIdent.Value
		} else {
			return nil, fmt.Errorf("invalid ROWS clause: must be a string selector")
		}

		doc.Find(rowsSelector).Each(func(i int, sel *goquery.Selection) {
			row := make(map[string]interface{})

			for _, f := range fields {
				var fieldSel *goquery.Selection
				if strings.HasPrefix(f.Selector, "+ ") {
					// Select next sibling and search within it
					siblingSel := strings.TrimSpace(strings.TrimPrefix(f.Selector, "+"))
					
					// Split sibling selector from the child selector if any
					// Example: "+ tr .score" -> sibling is "tr", child is ".score"
					parts := strings.SplitN(siblingSel, " ", 2)
					if len(parts) == 2 {
						fieldSel = sel.NextFiltered(parts[0]).Find(parts[1])
					} else {
						// e.g. "+ .subline"
						fieldSel = sel.NextFiltered(parts[0])
					}
				} else {
					fieldSel = sel.Find(f.Selector)
				}

				if fieldSel.Length() > 0 {
					val := strings.TrimSpace(fieldSel.First().Text())
					row[f.Name] = val
				} else {
					row[f.Name] = nil
				}
			}
			
			rows = append(rows, row)
		})
	} else {
		// Unscoped (legacy) extraction using zipping
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
	}

	// 5. WHERE clause
	if s.Where != nil {
		var filteredRows []map[string]interface{}
		for _, row := range rows {
			ok, err := e.evaluateWhere(s.Where, row)
			if err != nil {
				return nil, err
			}
			if ok {
				filteredRows = append(filteredRows, row)
			}
		}
		rows = filteredRows
	}

	return rows, nil
}

func (e *Evaluator) evaluateWhere(exp ast.Expression, row map[string]interface{}) (bool, error) {
	result, err := e.evaluateExpression(exp, row)
	if err != nil {
		return false, err
	}
	return isTruthy(result), nil
}

func (e *Evaluator) evaluateExpression(exp ast.Expression, row map[string]interface{}) (interface{}, error) {
	switch node := exp.(type) {
	case *ast.Identifier:
		if val, ok := row[node.Value]; ok {
			return val, nil
		}
		// If not found in row, it could be a raw CSS selector string
		return node.Value, nil
	case *ast.StringLiteral:
		return node.Value, nil
	case *ast.IntegerLiteral:
		return node.Value, nil
	case *ast.FloatLiteral:
		return node.Value, nil
	case *ast.Boolean:
		return node.Value, nil
	case *ast.NullLiteral:
		return nil, nil
	case *ast.InfixExpression:
		left, err := e.evaluateExpression(node.Left, row)
		if err != nil {
			return nil, err
		}
		right, err := e.evaluateExpression(node.Right, row)
		if err != nil {
			return nil, err
		}
		return evaluateInfix(node.Operator, left, right)
	case *ast.PrefixExpression:
		right, err := e.evaluateExpression(node.Right, row)
		if err != nil {
			return nil, err
		}
		return evaluatePrefix(node.Operator, right)
	}
	return nil, fmt.Errorf("unsupported expression type in WHERE: %T", exp)
}

func isTruthy(obj interface{}) bool {
	if obj == nil {
		return false
	}
	switch v := obj.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case float64:
		return v != 0.0
	case string:
		return v != ""
	}
	return true
}

var nonNumericRegex = regexp.MustCompile(`[^\d\.\-]+`)

func coerceToFloat(obj interface{}) (float64, error) {
	switch v := obj.(type) {
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		// Strip non-numeric
		cleanStr := nonNumericRegex.ReplaceAllString(v, "")
		if cleanStr == "" {
			return 0, fmt.Errorf("cannot coerce string to float: %s", v)
		}
		f, err := strconv.ParseFloat(cleanStr, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot coerce string to float: %s", v)
		}
		return f, nil
	}
	return 0, fmt.Errorf("cannot coerce to float: %v", obj)
}

func evaluateInfix(operator string, left, right interface{}) (interface{}, error) {
	switch operator {
	case "AND":
		return isTruthy(left) && isTruthy(right), nil
	case "OR":
		return isTruthy(left) || isTruthy(right), nil
	case "==", "=":
		lFloat, lErr := coerceToFloat(left)
		rFloat, rErr := coerceToFloat(right)
		if lErr == nil && rErr == nil {
			return lFloat == rFloat, nil
		}
		return left == right, nil
	case "!=":
		lFloat, lErr := coerceToFloat(left)
		rFloat, rErr := coerceToFloat(right)
		if lErr == nil && rErr == nil {
			return lFloat != rFloat, nil
		}
		return left != right, nil
	case ">", "<", ">=", "<=":
		l, err1 := coerceToFloat(left)
		r, err2 := coerceToFloat(right)
		if err1 != nil {
			return false, err1
		}
		if err2 != nil {
			return false, err2
		}
		switch operator {
		case ">":
			return l > r, nil
		case "<":
			return l < r, nil
		case ">=":
			return l >= r, nil
		case "<=":
			return l <= r, nil
		}
	}
	return nil, fmt.Errorf("unsupported operator: %s", operator)
}

func evaluatePrefix(operator string, right interface{}) (interface{}, error) {
	switch operator {
	case "!":
		return !isTruthy(right), nil
	}
	return nil, fmt.Errorf("unsupported prefix operator: %s", operator)
}
