package lexer

import (
	"fmt"
	"os"
	"strings"

	"language-betawi/internal/betawimsg"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte

	line   int
	column int

	Warnings []string
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	if l.ch == '/' && l.peekChar() == '/' {
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	l.skipComment()
	l.skipWhitespace()

	var tok Token
	tok.Line, tok.Column = l.line, l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: EQ, Literal: "==", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(ASSIGN, l.ch, tok.Line, tok.Column)
		}
	case '+':
		tok = newToken(PLUS, l.ch, tok.Line, tok.Column)
	case '-':
		tok = newToken(MINUS, l.ch, tok.Line, tok.Column)
	case '*':
		tok = newToken(ASTERISK, l.ch, tok.Line, tok.Column)
	case '/':
		tok = newToken(SLASH, l.ch, tok.Line, tok.Column)
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: NOT_EQ, Literal: "!=", Line: tok.Line, Column: tok.Column}
		} else {
			tok = newToken(BANG, l.ch, tok.Line, tok.Column)
		}
	case '<':
		tok = newToken(LT, l.ch, tok.Line, tok.Column)
	case '>':
		tok = newToken(GT, l.ch, tok.Line, tok.Column)
	case ',':
		tok = newToken(COMMA, l.ch, tok.Line, tok.Column)
	case ';':
		tok = newToken(SEMICOLON, l.ch, tok.Line, tok.Column)
	case '(':
		tok = newToken(LPAREN, l.ch, tok.Line, tok.Column)
	case ')':
		tok = newToken(RPAREN, l.ch, tok.Line, tok.Column)
	case '{':
		tok = newToken(LBRACE, l.ch, tok.Line, tok.Column)
	case '}':
		tok = newToken(RBRACE, l.ch, tok.Line, tok.Column)
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = EOF
		return tok
	default:
		if isLetter(l.ch) {
			word := l.readIdentifier()
			tok.Literal = word
			tok.Type = l.resolveWord(word, &tok)
			return tok
		} else if isDigit(l.ch) {
			numTok := l.readNumber()
			numTok.Line, numTok.Column = tok.Line, tok.Column
			return numTok
		}
		tok = newToken(ILLEGAL, l.ch, tok.Line, tok.Column)
	}

	l.readChar()
	return tok
}

func (l *Lexer) resolveWord(word string, tok *Token) TokenType {
	if tokType, ok := LookupExact(word); ok {
		return tokType
	}

	if result := LookupFuzzy(word); result.Matched {
		tok.FuzzyCorrected = true
		tok.OriginalWord = word
		tok.MatchScore = result.Score

		warning := betawimsg.FuzzyWarning(tok.Line, word, result.MatchedWord, result.Score*100)
		l.Warnings = append(l.Warnings, warning)
		fmt.Fprintln(os.Stderr, "⚠️  "+warning)

		return result.Type
	}

	return IDENT
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readNumber() Token {
	start := l.position
	tokType := INT
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokType = FLOAT
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return Token{Type: tokType, Literal: l.input[start:l.position]}
}

func (l *Lexer) readString() string {
	var sb strings.Builder
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
		sb.WriteByte(l.ch)
	}
	l.readChar()
	return sb.String()
}

func newToken(tokType TokenType, ch byte, line, col int) Token {
	return Token{Type: tokType, Literal: string(ch), Line: line, Column: col}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
