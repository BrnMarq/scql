package ast

import (
	"bytes"
	"scql/token"
	"strings"
)

// The base Node interface
type Node interface {
	TokenLiteral() string
	String() string
}

// All statement nodes implement this
type Statement interface {
	Node
	statementNode()
}

// All expression nodes implement this
type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// Statements

type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type OrderExpression struct {
	Token token.Token // ASC or DESC token
	Field Expression
	Order string
}

func (oe *OrderExpression) expressionNode()      {}
func (oe *OrderExpression) TokenLiteral() string { return oe.Token.Literal }
func (oe *OrderExpression) String() string {
	var out bytes.Buffer
	out.WriteString(oe.Field.String())
	out.WriteString(" ")
	out.WriteString(oe.Order)
	return out.String()
}

type SelectStatement struct {
	Token  token.Token // the 'SELECT' token
	Fields []Expression
	From   Expression
	Where  Expression
	Order  []*OrderExpression
	Rows   Expression
}

func (ss *SelectStatement) statementNode()       {}
func (ss *SelectStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SelectStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ss.TokenLiteral() + " ")

	fields := []string{}
	for _, f := range ss.Fields {
		fields = append(fields, f.String())
	}
	out.WriteString(strings.Join(fields, ", "))

	if ss.From != nil {
		out.WriteString(" FROM ")
		out.WriteString(ss.From.String())
	}

	if ss.Where != nil {
		out.WriteString(" WHERE ")
		out.WriteString(ss.Where.String())
	}

	if ss.Rows != nil {
		out.WriteString(" ROWS ")
		out.WriteString(ss.Rows.String())
	}

	if len(ss.Order) > 0 {
		out.WriteString(" ORDER BY ")
		orderFields := []string{}
		for _, of := range ss.Order {
			orderFields = append(orderFields, of.String())
		}
		out.WriteString(strings.Join(orderFields, ", "))
	}
	out.WriteString(";")

	return out.String()
}

type AuthenticateStatement struct {
	Token      token.Token // the 'AUTHENTICATE' token
	At         Expression
	SubmitForm bool
	WithParams []Expression
}

func (as *AuthenticateStatement) statementNode()       {}
func (as *AuthenticateStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AuthenticateStatement) String() string {
	var out bytes.Buffer

	out.WriteString(as.TokenLiteral() + " AT ")
	if as.At != nil {
		out.WriteString(as.At.String())
	}

	if as.SubmitForm {
		out.WriteString(" SUBMIT FORM")
	}

	if len(as.WithParams) > 0 {
		out.WriteString(" WITH (")
		params := []string{}
		for _, p := range as.WithParams {
			params = append(params, p.String())
		}
		out.WriteString(strings.Join(params, ", "))
		out.WriteString(")")
	}
	out.WriteString(";")

	return out.String()
}

type SetStatement struct {
	Token       token.Token // the 'SET' token
	Assignments []Expression
}

func (ss *SetStatement) statementNode()       {}
func (ss *SetStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ss.TokenLiteral() + " ")

	assignments := []string{}
	for _, a := range ss.Assignments {
		assignments = append(assignments, a.String())
	}
	out.WriteString(strings.Join(assignments, ", "))
	out.WriteString(";")

	return out.String()
}

// Expressions
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

type NullLiteral struct {
	Token token.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullLiteral) String() string       { return nl.Token.Literal }

type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (oe *InfixExpression) expressionNode()      {}
func (oe *InfixExpression) TokenLiteral() string { return oe.Token.Literal }
func (oe *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(oe.Left.String())
	out.WriteString(" " + oe.Operator + " ")
	out.WriteString(oe.Right.String())
	out.WriteString(")")

	return out.String()
}

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type AliasExpression struct {
	Token token.Token // The AS token
	Left  Expression
	Alias string
}

func (ae *AliasExpression) expressionNode()      {}
func (ae *AliasExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AliasExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ae.Left.String())
	out.WriteString(" AS ")
	out.WriteString(ae.Alias)

	return out.String()
}
