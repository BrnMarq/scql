package semantic

import (
	"scql/lexer"
	"scql/parser"
	"testing"
)

func TestSemanticAnalyzer_Valid(t *testing.T) {
	input := `
		SELECT ".title", "score" 
		FROM "example.com";
	`
	
	_, tokens := lexer.Lex("test", input)
	p := parser.New(tokens)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.Analyze(program)

	if len(analyzer.Errors()) != 0 {
		for _, err := range analyzer.Errors() {
			t.Errorf("Unexpected semantic error: %s", err)
		}
	}
}

func TestSemanticAnalyzer_InvalidURL(t *testing.T) {
	input := `
		SELECT id, non_existent_col
		FROM "not-a-valid-url";
	`
	
	_, tokens := lexer.Lex("test", input)
	p := parser.New(tokens)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.Analyze(program)

	errors := analyzer.Errors()
	if len(errors) == 0 {
		t.Fatalf("Expected semantic errors, got 0")
	}

	// We expect a URL validation error
	foundURLError := false
	for _, err := range errors {
		if err.Message == "FROM clause must be a valid URL, got 'https://not-a-valid-url'" {
			foundURLError = true
		}
	}
	
	if !foundURLError {
		t.Errorf("Expected URL validation error, got: %v", errors)
	}
}

func TestSemanticAnalyzer_InvalidFieldType(t *testing.T) {
	input := `
		SELECT 123, TRUE FROM "example.com";
	`
	
	_, tokens := lexer.Lex("test", input)
	p := parser.New(tokens)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.Analyze(program)

	errors := analyzer.Errors()
	if len(errors) != 2 {
		t.Fatalf("Expected 2 semantic errors, got %d: %v", len(errors), errors)
	}

	if errors[0].Message != "SELECT fields must be string CSS selectors or identifiers, got INT" {
		t.Errorf("Expected INT type error, got: %s", errors[0].Message)
	}

	if errors[1].Message != "SELECT fields must be string CSS selectors or identifiers, got BOOLEAN" {
		t.Errorf("Expected BOOLEAN type error, got: %s", errors[1].Message)
	}
}
