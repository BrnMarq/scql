package lexer

import "scanner/token"
import "unicode/utf8"
import "strings"
import "fmt"

type stateFn func(*Lexer) stateFn

type Lexer struct {
	name   string           // used only for error reports.
	input  string           // the string being scanned.
	start  int              // start position of this item.
	pos    int              // current position in the input.
	width  int              // width of last rune read from input.
	tokens chan token.Token // channel of scanned items.
	state  stateFn
}

const eof = -1

func (l *Lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.tokens) // No more tokens will be delivered.
}

func Lex(name, input string) (*Lexer, chan token.Token) {
	l := &Lexer{
		name:   name,
		input:  input,
		tokens: make(chan token.Token, 2),
	}
	go l.run() // Concurrently run state machine.
	return l, l.tokens
}

func (l *Lexer) emit(t token.TokenType) {
	l.tokens <- token.Token{Type: t, Literal: l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *Lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	var rune rune
	rune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return rune
}

func (l *Lexer) ignore() {
	l.start = l.pos
}

func (l *Lexer) backup() {
	l.pos -= l.width
}

func (l *Lexer) peek() rune {
	rune := l.next()
	l.backup()
	return rune
}

func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token.Token{
		Type:    token.ILLEGAL,
		Literal: fmt.Sprintf(format, args...),
	}
	return nil
}

func (l *Lexer) nextItem() token.Token {
	for {
		select {
		case token := <-l.tokens:
			return token
		default:
			l.state = l.state(l)
		}
	}
	panic("not reached")
}

// func (l *Lexer) NextToken() token.Token {
// 	var tok token.Token
//
// 	l.skipWhitespace()
//
// 	switch l.ch {
// 	case '=':
// 		tok = newToken(token.EQ, l.ch)
// 	case '+':
// 		tok = newToken(token.PLUS, l.ch)
// 	case '-':
// 		tok = newToken(token.MINUS, l.ch)
// 	case '!':
// 		if l.peekChar() == '=' {
// 			ch := l.ch
// 			l.readChar()
// 			literal := string(ch) + string(l.ch)
// 			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
// 		} else {
// 			tok = newToken(token.BANG, l.ch)
// 		}
// 	case '/':
// 		tok = newToken(token.SLASH, l.ch)
// 	case '*':
// 		tok = newToken(token.ASTERISK, l.ch)
// 	case '.':
// 		tok = newToken(token.DOT, l.ch)
// 	case '<':
// 		if l.peekChar() == '=' {
// 			ch := l.ch
// 			l.readChar()
// 			literal := string(ch) + string(l.ch)
// 			tok = token.Token{Type: token.LTE, Literal: literal}
// 		} else {
// 			tok = newToken(token.LT, l.ch)
// 		}
// 	case '>':
// 		if l.peekChar() == '=' {
// 			ch := l.ch
// 			l.readChar()
// 			literal := string(ch) + string(l.ch)
// 			tok = token.Token{Type: token.GTE, Literal: literal}
// 		} else {
// 			tok = newToken(token.GT, l.ch)
// 		}
// 	case ';':
// 		tok = newToken(token.SEMICOLON, l.ch)
// 	case ',':
// 		tok = newToken(token.COMMA, l.ch)
// 	case '(':
// 		tok = newToken(token.LPAREN, l.ch)
// 	case ')':
// 		tok = newToken(token.RPAREN, l.ch)
// 	case '\'':
// 		literal := l.readString()
// 		if literal == "" {
// 			tok = token.Token{Type: token.ILLEGAL, Literal: literal}
// 		} else {
// 			tok = token.Token{Type: token.STRING, Literal: literal}
// 		}
// 	case 0:
// 		tok.Literal = ""
// 		tok.Type = token.EOF
// 	default:
// 		if isLetter(l.ch) {
// 			tok.Literal = l.readIdentifier()
// 			tok.Type = token.LookupIdent(tok.Literal)
// 			return tok
// 		} else if isDigit(l.ch) {
// 			tok.Type = token.INT
// 			tok.Literal = l.readNumber()
// 			return tok
// 		} else {
// 			tok = newToken(token.ILLEGAL, l.ch)
// 		}
// 	}
//
// 	l.readChar()
// 	return tok
// }
//
// func (l *Lexer) skipWhitespace() {
// 	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
// 		l.readChar()
// 	}
// }
//
// func (l *Lexer) readChar() {
// 	if l.readPosition >= len(l.input) {
// 		l.ch = 0
// 	} else {
// 		l.ch = l.input[l.readPosition]
// 	}
// 	l.position = l.readPosition
// 	l.readPosition += 1
// }
//
// func (l *Lexer) peekChar() byte {
// 	if l.readPosition >= len(l.input) {
// 		return 0
// 	} else {
// 		return l.input[l.readPosition]
// 	}
// }
//
// func (l *Lexer) readIdentifier() string {
// 	position := l.position
// 	for isLetter(l.ch) {
// 		l.readChar()
// 	}
// 	return l.input[position:l.position]
// }
//
// func (l *Lexer) readString() string {
// 	position := l.position
// 	l.readChar()
// 	for l.ch != '\'' {
// 		if l.ch == 0 {
// 			return ""
// 		}
// 		l.readChar()
// 	}
// 	return l.input[position : l.position+1]
// }
//
// func (l *Lexer) readNumber() string {
// 	position := l.position
// 	for isDigit(l.ch) {
// 		l.readChar()
// 	}
// 	return l.input[position:l.position]
// }
//
// func isLetter(ch byte) bool {
// 	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
// }
//
// func isDigit(ch byte) bool {
// 	return '0' <= ch && ch <= '9'
// }
//
// func newToken(tokenType token.TokenType, ch byte) token.Token {
// 	return token.Token{Type: tokenType, Literal: string(ch)}
// }
