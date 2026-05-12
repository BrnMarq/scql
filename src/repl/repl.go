package repl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"scql/ast"
	"scql/evaluator"
	"scql/lexer"
	"scql/parser"
	"scql/semantic"
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

		fmt.Fprintln(out, "--- Lexer Output ---")
		_, tokens := lexer.Lex("Lexer", line)

		for token := range tokens {
			if token.Type == "EOF" {
				break
			}
			fmt.Fprintf(out, "{Type: %-15s Literal: %q}\n", token.Type, token.Literal)
		}

		fmt.Fprintln(out, "--- Parser Output ---")
		_, parserTokens := lexer.Lex("Lexer", line)
		p := parser.New(parserTokens)

		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		fmt.Fprintf(out, "Parsed String: %s\n", program.String())
		fmt.Fprintf(out, "Raw AST:\n%s\n", ast.PrintTree(program, "", true))

		fmt.Fprintln(out, "--- Semantic Analyzer Output ---")
		analyzer := semantic.NewAnalyzer()
		analyzer.Analyze(program)

		if len(analyzer.Errors()) > 0 {
			fmt.Fprintln(out, "Semantic Errors Found:")
			for _, err := range analyzer.Errors() {
				fmt.Fprintln(out, err.Error())
			}
			continue
		} else {
			fmt.Fprintln(out, "✅ Query is semantically valid!")
		}

		fmt.Fprintln(out, "--- Evaluator Output ---")
		eval := evaluator.NewEvaluator()
		results, err := eval.Evaluate(program)
		if err != nil {
			fmt.Fprintf(out, "Evaluator Error: %s\n", err)
			continue
		}

		b, _ := json.MarshalIndent(results, "", "  ")
		fmt.Fprintf(out, "%s\n", string(b))
	}
}

func Run(in io.Reader, out io.Writer) {
	bytes, err := io.ReadAll(in)
	if err != nil {
		fmt.Fprintf(out, "Error reading input: %s\n", err)
		return
	}
	input := string(bytes)

	fmt.Fprintln(out, "--- Lexer Output ---")
	_, tokens := lexer.Lex("Lexer", input)

	for token := range tokens {
		if token.Type == "EOF" {
			break
		}
		fmt.Fprintf(out, "{Type: %-15s Literal: %q}\n", token.Type, token.Literal)
	}

	fmt.Fprintln(out, "\n--- Parser Output ---")
	_, parserTokens := lexer.Lex("Lexer", input)
	p := parser.New(parserTokens)

	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParserErrors(out, p.Errors())
		return
	}

	fmt.Fprintf(out, "Parsed String:\n%s\n\n", program.String())
	fmt.Fprintf(out, "Raw AST:\n%s\n", ast.PrintTree(program, "", true))

	fmt.Fprintln(out, "--- Semantic Analyzer Output ---")
	analyzer := semantic.NewAnalyzer()
	analyzer.Analyze(program)

	if len(analyzer.Errors()) > 0 {
		fmt.Fprintln(out, "Semantic Errors Found:")
		for _, err := range analyzer.Errors() {
			fmt.Fprintln(out, err.Error())
		}
		return
	} else {
		fmt.Fprintln(out, "✅ Query is semantically valid!")
	}

	fmt.Fprintln(out, "--- Evaluator Output ---")
	eval := evaluator.NewEvaluator()
	results, evalErr := eval.Evaluate(program)
	if evalErr != nil {
		fmt.Fprintf(out, "Evaluator Error: %s\n", evalErr)
		return
	}

	b, _ := json.MarshalIndent(results, "", "  ")
	fmt.Fprintf(out, "%s\n", string(b))
}

func printParserErrors(out io.Writer, errors []string) {
	fmt.Fprintf(out, "Parser Errors:\n")
	for _, msg := range errors {
		fmt.Fprintf(out, "\t%s\n", msg)
	}
	fmt.Fprintln(out)
}
