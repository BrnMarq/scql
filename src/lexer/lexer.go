package lexer

import "scql/token"
import "unicode/utf8"
import "strings"
import "fmt"

type stateFn func(*Lexer) stateFn

type Lexer struct {
	name     string           // used only for error reports.
	input    string           // the string being scanned.
	start    int              // start position of this item.
	pos      int              // current position in the input.
	width    int              // width of last rune read from input.
	line     int              // current line number
	startCol int              // column at the start of the token
	col      int              // current column number
	tokens   chan token.Token // channel of scanned items.
	state    stateFn
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
		line:   1,
		col:    1,
	}
	go l.run() // Concurrently run state machine.
	return l, l.tokens
}

func (l *Lexer) updateLineCol() {
	for _, r := range l.input[l.start:l.pos] {
		if r == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
	}
}

func (l *Lexer) emit(t token.TokenType) {
	l.tokens <- token.Token{Type: t, Literal: l.input[l.start:l.pos], Line: l.line, Column: l.col}
	l.updateLineCol()
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
	l.updateLineCol()
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
		Line:    l.line,
		Column:  l.col,
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
