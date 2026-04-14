package repl

import (
	"bufio"
	"fmt"
	"io"
	"scanner/lexer"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		if line == "exit" || line == "\\q" {
			fmt.Fprintln(out, "Goodbye!")
			return
		}

		_, tokens := lexer.Lex("Lexer", line)

		for token := range tokens {
			if token.Type == "EOF" {
				break
			}
			fmt.Fprintf(out, "{Type: %-15s Literal: %q}\n", token.Type, token.Literal)
		}
	}
}
