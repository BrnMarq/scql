package parser

import (
	"scql/ast"
	"scql/lexer"
	"testing"
)

func TestSelectStatement(t *testing.T) {
	input := `SELECT * FROM "site.com" WHERE id = 5 ORDER BY price ASC, name DESC;`

	_, tokens := lexer.Lex("test", input)
	p := New(tokens)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("stmt is not ast.SelectStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(stmt.Fields))
	}
	testIdentifier(t, stmt.Fields[0], "*")

	if !testStringLiteral(t, stmt.From, "\"site.com\"") {
		return
	}

	if !testInfixExpression(t, stmt.Where, "id", "=", 5) {
		return
	}

	if len(stmt.Order) != 2 {
		t.Fatalf("expected 2 order expressions, got %d", len(stmt.Order))
	}
	testIdentifier(t, stmt.Order[0].Field, "price")
	if stmt.Order[0].Order != "ASC" {
		t.Errorf("expected order ASC, got %s", stmt.Order[0].Order)
	}
	testIdentifier(t, stmt.Order[1].Field, "name")
	if stmt.Order[1].Order != "DESC" {
		t.Errorf("expected order DESC, got %s", stmt.Order[1].Order)
	}
}

func TestAuthenticateStatement(t *testing.T) {
	input := `AUTHENTICATE AT "auth.com" SUBMIT FORM WITH (user="admin", pass="1234");`

	_, tokens := lexer.Lex("test", input)
	p := New(tokens)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.AuthenticateStatement)
	if !ok {
		t.Fatalf("stmt is not ast.AuthenticateStatement. got=%T", program.Statements[0])
	}

	testStringLiteral(t, stmt.At, "\"auth.com\"")

	if !stmt.SubmitForm {
		t.Errorf("expected SubmitForm to be true")
	}

	if len(stmt.WithParams) != 2 {
		t.Fatalf("expected 2 WITH params, got %d", len(stmt.WithParams))
	}

	testInfixExpression(t, stmt.WithParams[0], "user", "=", "\"admin\"")
	testInfixExpression(t, stmt.WithParams[1], "pass", "=", "\"1234\"")
}

func TestSetStatement(t *testing.T) {
	input := `SET a = TRUE, b = FALSE, c = NULL;`

	_, tokens := lexer.Lex("test", input)
	p := New(tokens)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SetStatement)
	if !ok {
		t.Fatalf("stmt is not ast.SetStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Assignments) != 3 {
		t.Fatalf("expected 3 assignments, got %d", len(stmt.Assignments))
	}

	testInfixExpression(t, stmt.Assignments[0], "a", "=", true)
	testInfixExpression(t, stmt.Assignments[1], "b", "=", false)
	testInfixExpression(t, stmt.Assignments[2], "c", "=", nil)
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"a AND b OR c",
			"((a AND b) OR c)",
		},
		{
			"a OR b AND c",
			"(a OR (b AND c))", // AND has higher precedence than OR
		},
		{
			"a OR b OR c",
			"((a OR b) OR c)",
		},
		{
			"a AND b AND c",
			"((a AND b) AND c)",
		},
		{
			"1 + 2 AND 3 + 4",
			"((1 + 2) AND (3 + 4))", // + has higher precedence than AND
		},
		{
			"1 < 2 OR 3 > 4",
			"((1 < 2) OR (3 > 4))", // < and > have higher precedence than OR
		},
	}

	for _, tt := range tests {
		_, tokens := lexer.Lex("test", tt.input)
		p := New(tokens)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

// testHelpers

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case float64:
		return testFloatLiteral(t, exp, v)
	case string:
		if _, ok := exp.(*ast.StringLiteral); ok {
			return testStringLiteral(t, exp, v)
		}
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	case nil:
		return testNullLiteral(t, exp)
	}
	t.Errorf("type of exp not handled. got=%T, expected=%v", exp, expected)
	return false
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	return true
}

func testFloatLiteral(t *testing.T, fl ast.Expression, value float64) bool {
	floatL, ok := fl.(*ast.FloatLiteral)
	if !ok {
		t.Errorf("fl not *ast.FloatLiteral. got=%T", fl)
		return false
	}

	if floatL.Value != value {
		t.Errorf("floatL.Value not %f. got=%f", value, floatL.Value)
		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}

	return true
}

func testStringLiteral(t *testing.T, exp ast.Expression, value string) bool {
	strL, ok := exp.(*ast.StringLiteral)
	if !ok {
		t.Errorf("exp not *ast.StringLiteral. got=%T", exp)
		return false
	}

	if strL.Value != value {
		t.Errorf("strL.Value not %s. got=%s", value, strL.Value)
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}

	return true
}

func testNullLiteral(t *testing.T, exp ast.Expression) bool {
	_, ok := exp.(*ast.NullLiteral)
	if !ok {
		t.Errorf("exp not *ast.NullLiteral. got=%T", exp)
		return false
	}
	return true
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.InfixExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}
