package lexer

import (
	"scanner/token"
	"testing"
)

func TestLexer_NextToken(t *testing.T) {
	input := `SELECT * FROM "www.example.com"` + "\n" +
		`WHERE class = 'title' AND id != 5 OR price <= 10.5` + "\n" +
		`SET a = TRUE, b = FALSE, c = NULL` + "\n" +
		`ORDER BY price ASC, name DESC` + "\n" +
		`AUTHENTICATE AT "auth.com" SUBMIT FORM WITH (user="admin", pass='1234')` + "\n" +
		`-- this is a line comment` + "\n" +
		`/* this is a /* nested */ block comment */` + "\n" +
		`! * / + - < > <= >= != = . ; , ( )` + "\n" +
		`123 45.67 _ident123`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.SELECT, "SELECT"},
		{token.ASTERISK, "*"},
		{token.FROM, "FROM"},
		{token.STRING, "\"www.example.com\""},
		{token.WHERE, "WHERE"},
		{token.IDENT, "class"},
		{token.EQ, "="},
		{token.STRING, "'title'"},
		{token.AND, "AND"},
		{token.IDENT, "id"},
		{token.NOT_EQ, "!="},
		{token.INT, "5"},
		{token.OR, "OR"},
		{token.IDENT, "price"},
		{token.LTE, "<="},
		{token.FLOAT, "10.5"},
		{token.SET, "SET"},
		{token.IDENT, "a"},
		{token.EQ, "="},
		{token.TRUE, "TRUE"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.EQ, "="},
		{token.FALSE, "FALSE"},
		{token.COMMA, ","},
		{token.IDENT, "c"},
		{token.EQ, "="},
		{token.NULL, "NULL"},
		{token.ORDER, "ORDER"},
		{token.BY, "BY"},
		{token.IDENT, "price"},
		{token.ASC, "ASC"},
		{token.COMMA, ","},
		{token.IDENT, "name"},
		{token.DESC, "DESC"},
		{token.AUTHENTICATE, "AUTHENTICATE"},
		{token.AT, "AT"},
		{token.STRING, "\"auth.com\""},
		{token.SUBMIT, "SUBMIT"},
		{token.FORM, "FORM"},
		{token.WITH, "WITH"},
		{token.LPAREN, "("},
		{token.IDENT, "user"},
		{token.EQ, "="},
		{token.STRING, "\"admin\""},
		{token.COMMA, ","},
		{token.IDENT, "pass"},
		{token.EQ, "="},
		{token.STRING, "'1234'"},
		{token.RPAREN, ")"},
		{token.COMMENT, "-- this is a line comment\n"},
		{token.BLOCK_COMMENT, "/* this is a /* nested */ block comment */"},
		{token.BANG, "!"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.LTE, "<="},
		{token.GTE, ">="},
		{token.NOT_EQ, "!="},
		{token.EQ, "="},
		{token.DOT, "."},
		{token.SEMICOLON, ";"},
		{token.COMMA, ","},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.INT, "123"},
		{token.FLOAT, "45.67"},
		{token.IDENT, "_ident123"},
		{token.EOF, ""},
	}

	_, tokens := Lex("test", input)

	for i, tt := range tests {
		tok := <-tokens

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q, literal=%q",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexerErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "Unterminated Double Quote String",
			input:         `"this string never ends`,
			expectedError: "Missing closing quote on string",
		},
		{
			name:          "Unterminated Single Quote String",
			input:         `'this string never ends`,
			expectedError: "Missing closing quote on string",
		},
		{
			name:          "Unterminated Block Comment",
			input:         `/* this is a /* nested */ comment that never ends`,
			expectedError: "Missing closing block comment",
		},
		{
			name:          "Illegal Character",
			input:         `SELECT @ FROM`,
			expectedError: "illegal character U+0040 '@'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, tokens := Lex("test_error", tt.input)

			foundIllegal := false
			for tok := range tokens {
				if tok.Type == token.ILLEGAL {
					foundIllegal = true
					if tok.Literal != tt.expectedError {
						t.Errorf("Expected error message %q, but got %q", tt.expectedError, tok.Literal)
					}
					break
				}
				if tok.Type == token.EOF {
					break
				}
			}

			if !foundIllegal {
				t.Errorf("Expected ILLEGAL token for input %q, but got none", tt.input)
			}
		})
	}
}
