package asyncpi

//go:generate go tool yacc -p asyncpi -o parser.y.go asyncpi.y

import "io"

// Lexer for asyncpi.
type Lexer struct {
	scanner *Scanner
	Errors  chan error
}

// NewLexer returns a new yacc-compatible lexer.
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{scanner: NewScanner(r), Errors: make(chan error, 1)}
}

// Lex is provided for yacc-compatible parser.
func (l *Lexer) Lex(yylval *asyncpiSymType) int {
	var token Token
	token, yylval.strval, _, _ = l.scanner.Scan()
	return int(token)
}

// Error handles error.
func (l *Lexer) Error(err string) {
	l.Errors <- &ErrParse{Err: err, Pos: l.scanner.pos}
}
