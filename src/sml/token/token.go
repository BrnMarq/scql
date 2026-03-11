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
	DELETE = "DELETE"
	INSERT = "INSERT"
	UPDATE = "UPDATE"

	// Target Keywords
	INTO   = "INTO"
	VALUES = "VALUES"
	FROM   = "FROM"
	WHERE  = "WHERE"
	SET    = "SET"

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
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	SELECT: "SELECT",
	DELETE: "DELETE",
	INSERT: "INSERT",
	UPDATE: "UPDATE",
	INTO:   "INTO",
	VALUES: "VALUES",
	FROM:   "FROM",
	WHERE:  "WHERE",
	SET:    "SET",
	OR:     "OR",
	AND:    "AND",
	TRUE:   "TRUE",
	FALSE:  "FALSE",
	NULL:   "NULL",
	ORDER:  "ORDER",
	BY:     "BY",
	ASC:    "ASC",
	DESC:   "DESC",
}

func LookupIdent(ident string) TokenType {
	lowercase_identifier := strings.ToUpper(ident)
	if tok, ok := keywords[lowercase_identifier]; ok {
		return tok
	}

	return IDENT
}
