package token

import "strings"

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT  = "IDENT"
	INT    = "INT"
	FLOAT  = "FLOAT"
	STRING = "STRING"

	// Operators
	EQ       = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	DOT      = "."

	LT  = "<"
	LTE = "<="
	GT  = ">"
	GTE = ">="

	NOT_EQ = "!="

	// Comments
	COMMENT       = "COMMENT"
	BLOCK_COMMENT = "BLOCK_COMMENT"

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"

	// Query Keywords
	SELECT = "SELECT"

	// Target Keywords
	FROM  = "FROM"
	WHERE = "WHERE"
	SET   = "SET"

	// Union and Intersection operators
	OR  = "OR"
	AND = "AND"

	// Value Keywords
	TRUE  = "TRUE"
	FALSE = "FALSE"
	NULL  = "NULL"

	// Order Keywords
	ORDER = "ORDER"
	BY    = "BY"
	ASC   = "ASC"
	DESC  = "DESC"
	AS    = "AS"

	// Auth Keywords
	AUTHENTICATE = "AUTHENTICATE"
	AT           = "AT"
	SUBMIT       = "SUBMIT"
	FORM         = "FORM"
	WITH         = "WITH"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	SELECT:       "SELECT",
	FROM:         "FROM",
	WHERE:        "WHERE",
	SET:          "SET",
	OR:           "OR",
	AND:          "AND",
	TRUE:         "TRUE",
	FALSE:        "FALSE",
	NULL:         "NULL",
	ORDER:        "ORDER",
	BY:           "BY",
	ASC:          "ASC",
	DESC:         "DESC",
	AS:           "AS",
	AUTHENTICATE: "AUTHENTICATE",
	AT:           "AT",
	SUBMIT:       "SUBMIT",
	FORM:         "FORM",
	WITH:         "WITH",
}

func LookupIdent(ident string) TokenType {
	uppercase_identifier := strings.ToUpper(ident)
	if tok, ok := keywords[uppercase_identifier]; ok {
		return tok
	}

	return IDENT
}
