package asyncpi

import "fmt"

// Tokens for use with lexer and parser.

// Token is a lexical token.
type Token int

const (
	// ILLEGAL is a special token for errors.
	ILLEGAL Token = iota
)

var eof = rune(0)

// TokenPos is a pair of coordinate to identify start of token.
type TokenPos struct {
	Char  int
	Lines []int
}

func (p TokenPos) String() string {
	return fmt.Sprintf("%d:%d", len(p.Lines)+1, p.Char)
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isAlphaNum(ch rune) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ('0' <= ch && ch <= '9')
}

func isNameSymbols(ch rune) bool {
	return ch == '_' || ch == '-'
}
