package ast

import (
	"scql/token"
	"testing"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&SelectStatement{
				Token: token.Token{Type: token.SELECT, Literal: "SELECT"},
				Fields: []Expression{
					&Identifier{
						Token: token.Token{Type: token.IDENT, Literal: "name"},
						Value: "name",
					},
				},
				From: &StringLiteral{
					Token: token.Token{Type: token.STRING, Literal: "site.com"},
					Value: "site.com",
				},
				Where: &InfixExpression{
					Token:    token.Token{Type: token.EQ, Literal: "="},
					Left:     &Identifier{Token: token.Token{Type: token.IDENT, Literal: "id"}, Value: "id"},
					Operator: "=",
					Right:    &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
				},
				Order: []*OrderExpression{
					{
						Token: token.Token{Type: token.DESC, Literal: "DESC"},
						Field: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "price"}, Value: "price"},
						Order: "DESC",
					},
				},
			},
		},
	}

	expected := "SELECT name FROM site.com WHERE (id = 5) ORDER BY price DESC;"
	if program.String() != expected {
		t.Errorf("program.String() wrong. expected=%q, got=%q", expected, program.String())
	}
}
