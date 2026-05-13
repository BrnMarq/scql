package parser

import (
	"fmt"
	"scql/ast"
	"scql/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	ALIAS       // AS
	OR          // OR
	AND         // AND
	EQUALS      // ==, =, !=
	LESSGREATER // >, <, >=, <=
	SUM         // +, -
	PRODUCT     // *, /
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

var precedences = map[token.TokenType]int{
	token.OR:       OR,
	token.AND:      AND,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LTE:      LESSGREATER,
	token.GTE:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.AS:       ALIAS,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	tokens chan token.Token
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(tokens chan token.Token) *Parser {
	p := &Parser{
		tokens: tokens,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.NULL, p.parseNullLiteral)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.ASTERISK, p.parseIdentifier)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)

	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.AS, p.parseAliasExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	for {
		p.peekToken = <-p.tokens
		if p.peekToken.Type != token.COMMENT && p.peekToken.Type != token.BLOCK_COMMENT {
			break
		}
	}
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.SELECT:
		return p.parseSelectStatement()
	case token.AUTHENTICATE:
		return p.parseAuthenticateStatement()
	case token.SET:
		return p.parseSetStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseSelectStatement() *ast.SelectStatement {
	stmt := &ast.SelectStatement{Token: p.curToken}

	p.nextToken() // past SELECT

	// Parse Fields
	stmt.Fields = []ast.Expression{}
	if p.curTokenIs(token.ASTERISK) {
		stmt.Fields = append(stmt.Fields, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	} else {
		stmt.Fields = append(stmt.Fields, p.parseExpression(LOWEST))
		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // to comma
			p.nextToken() // past comma
			stmt.Fields = append(stmt.Fields, p.parseExpression(LOWEST))
		}
	}

	if p.peekTokenIs(token.FROM) {
		p.nextToken() // to FROM
		p.nextToken() // past FROM
		stmt.From = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.WHERE) {
		p.nextToken() // to WHERE
		p.nextToken() // past WHERE
		stmt.Where = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.ROWS) {
		p.nextToken() // to ROWS
		p.nextToken() // past ROWS
		stmt.Rows = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.ORDER) {
		p.nextToken() // to ORDER
		if !p.expectPeek(token.BY) {
			return nil
		}
		p.nextToken() // past BY
		stmt.Order = p.parseOrderExpressions()
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseOrderExpressions() []*ast.OrderExpression {
	orders := []*ast.OrderExpression{}

	orders = append(orders, p.parseOrderExpression())

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // to comma
		p.nextToken() // past comma
		orders = append(orders, p.parseOrderExpression())
	}
	return orders
}

func (p *Parser) parseOrderExpression() *ast.OrderExpression {
	field := p.parseExpression(LOWEST)

	order := "ASC"
	var tok token.Token
	if p.peekTokenIs(token.ASC) {
		p.nextToken()
		tok = p.curToken
		order = "ASC"
	} else if p.peekTokenIs(token.DESC) {
		p.nextToken()
		tok = p.curToken
		order = "DESC"
	} else {
		tok = token.Token{Type: token.ASC, Literal: "ASC"}
	}

	return &ast.OrderExpression{Token: tok, Field: field, Order: order}
}

func (p *Parser) parseAuthenticateStatement() *ast.AuthenticateStatement {
	stmt := &ast.AuthenticateStatement{Token: p.curToken}

	if !p.expectPeek(token.AT) {
		return nil
	}
	p.nextToken()

	stmt.At = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SUBMIT) {
		p.nextToken() // to SUBMIT
		if !p.expectPeek(token.FORM) {
			return nil
		}
		stmt.SubmitForm = true
	}

	if p.peekTokenIs(token.WITH) {
		p.nextToken() // to WITH
		if !p.expectPeek(token.LPAREN) {
			return nil
		}
		stmt.WithParams = p.parseCallArguments()
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseSetStatement() *ast.SetStatement {
	stmt := &ast.SetStatement{Token: p.curToken}

	p.nextToken() // past SET

	stmt.Assignments = []ast.Expression{}
	stmt.Assignments = append(stmt.Assignments, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // to comma
		p.nextToken() // past comma
		stmt.Assignments = append(stmt.Assignments, p.parseExpression(LOWEST))
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	// For actual implementation, stripping quotes might be needed if token literal includes them
	// Depending on lexer implementation, it might contain double quotes or not.
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseAliasExpression(left ast.Expression) ast.Expression {
	exp := &ast.AliasExpression{
		Token: p.curToken,
		Left:  left,
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	exp.Alias = p.curToken.Literal
	return exp
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
