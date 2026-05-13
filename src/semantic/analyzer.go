package semantic

import (
	"fmt"
	"net/url"
	"scql/ast"
	"scql/token"
	"strings"
)

type SemanticError struct {
	Line    int
	Column  int
	Message string
}

func (e *SemanticError) Error() string {
	return fmt.Sprintf("[Line %d:%d] Semantic Error: %s", e.Line, e.Column, e.Message)
}

type Analyzer struct {
	errors []*SemanticError

	// Indicates whether we are inside a SELECT statement fetching a URL
	inSelectContext bool
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		errors: []*SemanticError{},
	}
}

func (a *Analyzer) Errors() []*SemanticError {
	return a.errors
}

func (a *Analyzer) reportError(tok token.Token, format string, args ...interface{}) {
	a.errors = append(a.errors, &SemanticError{
		Line:    tok.Line,
		Column:  tok.Column,
		Message: fmt.Sprintf(format, args...),
	})
}

// Analyze traverses the AST and performs semantic checks.
func (a *Analyzer) Analyze(program *ast.Program) {
	for _, stmt := range program.Statements {
		a.analyzeStatement(stmt)
	}
}

func stripQuotes(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[0] == s[len(s)-1] {
		return s[1 : len(s)-1]
	}
	return s
}

func (a *Analyzer) analyzeStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		a.analyzeExpression(s.Expression)
	case *ast.SelectStatement:
		a.analyzeSelectStatement(s)
	case *ast.SetStatement:
		a.analyzeSetStatement(s)
	case *ast.AuthenticateStatement:
		a.analyzeAuthenticateStatement(s)
	default:
		// Not strictly an error, just unhandled
	}
}

func (a *Analyzer) analyzeSelectStatement(s *ast.SelectStatement) {
	a.inSelectContext = true
	defer func() { a.inSelectContext = false }()

	// First analyze FROM clause to validate URL
	if s.From != nil {
		var targetURL string
		if fromString, ok := s.From.(*ast.StringLiteral); ok {
			targetURL = stripQuotes(fromString.Value)
		} else if fromIdent, ok := s.From.(*ast.Identifier); ok {
			targetURL = stripQuotes(fromIdent.Value)
		} else {
			a.reportError(s.Token, "FROM clause must be a URL string or identifier")
			return
		}

		if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
			targetURL = "https://" + targetURL
		}

		u, err := url.ParseRequestURI(targetURL)
		if err != nil || u.Host == "" || !strings.Contains(u.Host, ".") {
			a.reportError(s.Token, "FROM clause must be a valid URL, got '%s'", targetURL)
		}
	}

	// Analyze fields (SELECT columns)
	for _, field := range s.Fields {
		fieldType := a.analyzeExpression(field)
		if fieldType != TypeString && fieldType != TypeUnknown {
			a.reportError(s.Token, "SELECT fields must be string CSS selectors or identifiers, got %s", fieldType)
		}
	}

	// Analyze WHERE
	if s.Where != nil {
		whereType := a.analyzeExpression(s.Where)
		if whereType != TypeBoolean && whereType != TypeUnknown {
			a.reportError(s.Token, "WHERE clause must evaluate to a boolean expression, got %s", whereType)
		}
	}

	// Analyze ROWS
	if s.Rows != nil {
		rowsType := a.analyzeExpression(s.Rows)
		if rowsType != TypeString && rowsType != TypeUnknown {
			a.reportError(s.Token, "ROWS clause must evaluate to a string CSS selector, got %s", rowsType)
		}
	}

	// Analyze ORDER BY
	for _, order := range s.Order {
		a.analyzeExpression(order.Field)
	}
}

func (a *Analyzer) analyzeSetStatement(s *ast.SetStatement) {
	for _, assignment := range s.Assignments {
		if infix, ok := assignment.(*ast.InfixExpression); ok && infix.Operator == "=" {
			if _, ok := infix.Left.(*ast.Identifier); !ok {
				a.reportError(infix.Token, "left side of assignment must be an identifier")
			}
			// Just analyzing both sides to catch any expression errors
			a.analyzeExpression(infix.Right)
		} else {
			// Assume it's a syntax error caught by parser, but just in case
			tok := token.Token{Line: 0, Column: 0}
			a.reportError(tok, "SET statement expects assignments in the form of 'variable = expression'")
		}
	}
}

func (a *Analyzer) analyzeAuthenticateStatement(s *ast.AuthenticateStatement) {
	if s.At != nil {
		atType := a.analyzeExpression(s.At)
		if atType != TypeString && atType != TypeUnknown {
			a.reportError(s.Token, "AUTHENTICATE AT must be a string URL, got %s", atType)
		}
	}

	for _, param := range s.WithParams {
		if infix, ok := param.(*ast.InfixExpression); ok && infix.Operator == "=" {
			a.analyzeExpression(infix.Right)
		} else {
			tok := token.Token{Line: 0, Column: 0}
			a.reportError(tok, "AUTHENTICATE WITH expects assignments in the form of 'key = value'")
		}
	}
}

// analyzeExpression returns the resulting Type of an expression.
func (a *Analyzer) analyzeExpression(exp ast.Expression) Type {
	switch e := exp.(type) {
	case *ast.IntegerLiteral:
		return TypeInt
	case *ast.FloatLiteral:
		return TypeFloat
	case *ast.StringLiteral:
		return TypeString
	case *ast.Boolean:
		return TypeBoolean
	case *ast.NullLiteral:
		return TypeNull
	case *ast.Identifier:
		// In scql, identifiers in a SELECT clause represent string CSS selectors.
		// Variables in SET clauses don't have strict type tracking anymore since
		// we removed the mock schema.
		return TypeString

	case *ast.PrefixExpression:
		rightType := a.analyzeExpression(e.Right)
		switch e.Operator {
		case "!":
			if rightType != TypeBoolean && rightType != TypeUnknown {
				a.reportError(e.Token, "operator '!' requires boolean operand, got %s", rightType)
			}
			return TypeBoolean
		case "-":
			if rightType != TypeInt && rightType != TypeFloat && rightType != TypeUnknown {
				a.reportError(e.Token, "operator '-' requires numeric operand, got %s", rightType)
			}
			return rightType
		}

	case *ast.InfixExpression:
		leftType := a.analyzeExpression(e.Left)
		rightType := a.analyzeExpression(e.Right)

		if leftType == TypeUnknown || rightType == TypeUnknown {
			return TypeUnknown // Cascade the unknown type to prevent spamming errors
		}

		switch e.Operator {
		case "+", "-", "*", "/":
			if !isNumericOrString(leftType) || !isNumericOrString(rightType) {
				a.reportError(e.Token, "type mismatch: operator '%s' requires numeric types (or coercible strings), got %s and %s", e.Operator, leftType, rightType)
				return TypeUnknown
			}
			if leftType == TypeFloat || rightType == TypeFloat {
				return TypeFloat
			}
			return TypeInt
		case "<", ">", "<=", ">=":
			if !isNumericOrString(leftType) || !isNumericOrString(rightType) {
				a.reportError(e.Token, "type mismatch: operator '%s' requires numeric types (or coercible strings), got %s and %s", e.Operator, leftType, rightType)
			}
			return TypeBoolean
		case "==", "!=", "=":
			if leftType != rightType {
				// Allow comparing anything to NULL, and strings to numbers
				if leftType != TypeNull && rightType != TypeNull && !(isNumericOrString(leftType) && isNumericOrString(rightType)) {
					a.reportError(e.Token, "type mismatch: cannot compare %s and %s", leftType, rightType)
				}
			}
			return TypeBoolean
		case "AND", "OR":
			if leftType != TypeBoolean || rightType != TypeBoolean {
				a.reportError(e.Token, "type mismatch: operator '%s' requires boolean types, got %s and %s", e.Operator, leftType, rightType)
			}
			return TypeBoolean
		}

	case *ast.CallExpression:
		for _, arg := range e.Arguments {
			a.analyzeExpression(arg)
		}
		return TypeUnknown

	case *ast.AliasExpression:
		// The type of the alias is the type of the expression it's aliasing
		return a.analyzeExpression(e.Left)
	}

	return TypeUnknown
}

func isNumeric(t Type) bool {
	return t == TypeInt || t == TypeFloat
}

func isNumericOrString(t Type) bool {
	return isNumeric(t) || t == TypeString
}
