package ast_test

import (
	"fmt"
	"scql/ast"
	"scql/lexer"
	"scql/parser"
	"testing"
)

func TestPrintTree(t *testing.T) {
	input1 := `SELECT title, price FROM "https://example.com" WHERE price < 100 ORDER BY price ASC;`
	input2 := `AUTHENTICATE AT "https://api.example.com" SUBMIT FORM WITH ("user", "pass");`
	input3 := `SET "Authorization", "Bearer token";`

	for i, input := range []string{input1, input2, input3} {
		_, parserTokens := lexer.Lex("Lexer", input)
		p := parser.New(parserTokens)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			t.Fatalf("Parser errors for input %d: %v", i+1, p.Errors())
		}

		fmt.Printf("--- Test %d ---\n", i+1)
		fmt.Println(ast.PrintTree(program, "", true))
	}
}
