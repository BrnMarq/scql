package lexer

import (
	"scanner/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `val x = 5
val y = 10
val f = fn x => x + 10
val s = "string"
val c = #"c"
val b = true
val bf = false
val u = ()
val r = 3.14
val n = ~5
[1, 2]
(1, 2)
1 :: 2
x : int
[1] @ [2]
1 + 2 * 3 / 4 div 5 mod 6 ^ 7
x = y
x <> y
x < y
x > y
x <= y
x >= y
x andalso y
x orelse y
not x
if x then y else z
fun add a b = a + b
case x of
  | 1 => "one"
  | _ => "other"
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.VAL, "val"},
		{token.IDENTIFIER, "x"},
		{token.EQUAL, "="},
		{token.INTEGER, "5"},
		{token.VAL, "val"},
		{token.IDENTIFIER, "y"},
		{token.EQUAL, "="},
		{token.INTEGER, "10"},
		{token.VAL, "val"},
		{token.IDENTIFIER, "f"},
		{token.EQUAL, "="},
		{token.IDENTIFIER, "fn"}, // fn is not a keyword in token.go, so it's an ident
		{token.IDENTIFIER, "x"},
		{token.DOUBLE_ARROW, "=>"},
		{token.IDENTIFIER, "x"},
		{token.PLUS, "+"},
		{token.INTEGER, "10"},
		{token.VAL, "val"},
		{token.IDENTIFIER, "s"},
		{token.EQUAL, "="},
		{token.STRING, "\"string\""},
		{token.VAL, "val"},
		{token.IDENTIFIER, "c"},
		{token.EQUAL, "="},
		{token.CHAR, "#\"c\""},
		{token.VAL, "val"},
		{token.IDENTIFIER, "b"},
		{token.EQUAL, "="},
		{token.BOOL, "true"},
		{token.VAL, "val"},
		{token.IDENTIFIER, "bf"},
		{token.EQUAL, "="},
		{token.BOOL, "false"},
		{token.VAL, "val"},
		{token.IDENTIFIER, "u"},
		{token.EQUAL, "="},
		{token.UNIT, "()"},
		{token.VAL, "val"},
		{token.IDENTIFIER, "r"},
		{token.EQUAL, "="},
		{token.REAL, "3.14"},
		{token.VAL, "val"},
		{token.IDENTIFIER, "n"},
		{token.EQUAL, "="},
		{token.INTEGER, "~5"},
		{token.LBRACKET, "["},
		{token.INTEGER, "1"},
		{token.COMMA, ","},
		{token.INTEGER, "2"},
		{token.RBRACKET, "]"},
		{token.LPAREN, "("},
		{token.INTEGER, "1"},
		{token.COMMA, ","},
		{token.INTEGER, "2"},
		{token.RPAREN, ")"},
		{token.INTEGER, "1"},
		{token.DOUBLE_COLON, "::"},
		{token.INTEGER, "2"},
		{token.IDENTIFIER, "x"},
		{token.COLON, ":"},
		{token.IDENTIFIER, "int"},
		{token.LBRACKET, "["},
		{token.INTEGER, "1"},
		{token.RBRACKET, "]"},
		{token.AT, "@"},
		{token.LBRACKET, "["},
		{token.INTEGER, "2"},
		{token.RBRACKET, "]"},
		{token.INTEGER, "1"},
		{token.PLUS, "+"},
		{token.INTEGER, "2"},
		{token.ASTERISK, "*"},
		{token.INTEGER, "3"},
		{token.SLASH, "/"},
		{token.INTEGER, "4"},
		{token.DIVISION, "div"},
		{token.INTEGER, "5"},
		{token.MODULO, "mod"},
		{token.INTEGER, "6"},
		{token.CARET, "^"},
		{token.INTEGER, "7"},
		{token.IDENTIFIER, "x"},
		{token.EQUAL, "="},
		{token.IDENTIFIER, "y"},
		{token.IDENTIFIER, "x"},
		{token.NOT_EQUAL, "<>"},
		{token.IDENTIFIER, "y"},
		{token.IDENTIFIER, "x"},
		{token.LT, "<"},
		{token.IDENTIFIER, "y"},
		{token.IDENTIFIER, "x"},
		{token.GT, ">"},
		{token.IDENTIFIER, "y"},
		{token.IDENTIFIER, "x"},
		{token.LTE, "<="},
		{token.IDENTIFIER, "y"},
		{token.IDENTIFIER, "x"},
		{token.GTE, ">="},
		{token.IDENTIFIER, "y"},
		{token.IDENTIFIER, "x"},
		{token.ANDALSO, "andalso"},
		{token.IDENTIFIER, "y"},
		{token.IDENTIFIER, "x"},
		{token.ORELSE, "orelse"},
		{token.IDENTIFIER, "y"},
		{token.NOT, "not"},
		{token.IDENTIFIER, "x"},
		{token.IF, "if"},
		{token.IDENTIFIER, "x"},
		{token.THEN, "then"},
		{token.IDENTIFIER, "y"},
		{token.ELSE, "else"},
		{token.IDENTIFIER, "z"},
		{token.FUN, "fun"},
		{token.IDENTIFIER, "add"},
		{token.IDENTIFIER, "a"},
		{token.IDENTIFIER, "b"},
		{token.EQUAL, "="},
		{token.IDENTIFIER, "a"},
		{token.PLUS, "+"},
		{token.IDENTIFIER, "b"},
		{token.CASE, "case"},
		{token.IDENTIFIER, "x"},
		{token.OF, "of"},
		{token.PIPE, "|"},
		{token.INTEGER, "1"},
		{token.DOUBLE_ARROW, "=>"},
		{token.STRING, "\"one\""},
		{token.PIPE, "|"},
		{token.IDENTIFIER, "_"},
		{token.DOUBLE_ARROW, "=>"},
		{token.STRING, "\"other\""},
		{token.EOF, ""},
	}

	_, tokens := Lex("test", input)

	for i, tt := range tests {
		tok := <-tokens

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
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
		expectedError token.TokenType
	}{
		{
			name:          "Unterminated String",
			input:         `"this string never ends`,
			expectedError: token.ILLEGAL,
		},
		{
			name:          "Invalid Char Syntax",
			input:         `#i`,
			expectedError: token.ILLEGAL,
		},
		{
			name:          "Unterminated Char",
			input:         `#"c`,
			expectedError: token.ILLEGAL,
		},
		{
			name:          "Invalid Negative Number",
			input:         `~a`,
			expectedError: token.ILLEGAL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, tokens := Lex("test_error", tt.input)

			foundIllegal := false
			for tok := range tokens {
				if tok.Type == token.ILLEGAL {
					foundIllegal = true
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
