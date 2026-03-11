package lexer

import "scanner/token"
import "unicode"

func lexStart(l *Lexer) stateFn {
	for {
		switch r := l.next(); {
		case isSpace(r):
			l.ignore()
		case r == eof:
			l.emit(token.EOF)
			return nil
		case r == '=':
			l.emit(token.EQ)
		case r == '+':
			l.emit(token.PLUS)
		case r == '-':
			if l.peek() == '-' {
				l.next()
				return lexComment
			} else {
				l.emit(token.MINUS)
			}
		case r == '!':
			if l.peek() == '=' {
				l.next()
				l.emit(token.NOT_EQ)
			} else {
				l.emit(token.BANG)
			}
		case r == '/':
			if l.peek() == '*' {
				l.next()
				return lexBlockComment
			}
			l.emit(token.SLASH)
		case r == '*':
			l.emit(token.ASTERISK)
		case r == '.':
			l.emit(token.DOT)
		case r == '<':
			if l.peek() == '=' {
				l.next()
				l.emit(token.LTE)
			} else {
				l.emit(token.LT)
			}
		case r == '>':
			if l.peek() == '=' {
				l.next()
				l.emit(token.GTE)
			} else {
				l.emit(token.GT)
			}
		case r == ';':
			l.emit(token.SEMICOLON)
		case r == ',':
			l.emit(token.COMMA)
		case r == '(':
			l.emit(token.LPAREN)
		case r == ')':
			l.emit(token.RPAREN)
		case r == '\'':
			return lexString
		default:
			if unicode.IsLetter(r) || r == '_' {
				l.backup()
				return lexLetter
			} else if unicode.IsDigit(r) {
				l.backup()
				return lexNumber
			}
		}
	}
}

func lexComment(l *Lexer) stateFn {
	for r := l.next(); r != '\n' && r != eof; r = l.next() {
	}
	l.emit(token.COMMENT)
	return lexStart
}

func lexBlockComment(l *Lexer) stateFn {
	counter := 1
	for r := l.next(); counter > 0; r = l.next() {
		if r == '/' {
			if l.peek() == '*' {
				counter += 1
			}
		} else if r == '*' {
			if l.peek() == '/' {
				counter -= 1
			}
		} else if r == eof {
			return l.errorf("Missing closing block comment")
		}
	}
	l.emit(token.BLOCK_COMMENT)
	return lexStart
}

func lexString(l *Lexer) stateFn {
	for r := l.next(); r != '\''; r = l.next() {
		if r == eof {
			return l.errorf("Missing closing double quotes on string")
		}
	}
	l.emit(token.STRING)
	return lexStart
}

func lexLetter(l *Lexer) stateFn {
	for r := l.next(); isAlphaNumeric(r); r = l.next() {

	}
	l.backup()
	word := l.input[l.start:l.pos]
	l.emit(token.LookupIdent(word))
	return lexStart
}

func lexNumber(l *Lexer) stateFn {
	isFloat := false
	// Optional leading sign.
	l.accept("+~")
	digits := "0123456789"
	l.acceptRun(digits)
	if l.accept(".") {
		isFloat = true
		l.acceptRun(digits)
	}
	if isFloat {
		l.emit(token.FLOAT)
	} else {
		l.emit(token.INT)
	}
	return lexStart
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func isAlphaNumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
