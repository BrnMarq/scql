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
		_, tokens := lexer.Lex("Lexer", line)

		for token := range tokens {
			fmt.Printf("%+v\n", token)
		}
	}
}
