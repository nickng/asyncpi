package asyncpi

//go:generate goyacc -p asyncpi -o parser.y.go asyncpi.y

import "io"

// lexer for asyncpi.
type lexer struct {
	scanner *scanner
	Errors  chan error
}

// newLexer returns a new yacc-compatible lexer.
func newLexer(r io.Reader) *lexer {
	return &lexer{scanner: newScanner(r), Errors: make(chan error, 1)}
}

// Lex is provided for yacc-compatible parser.
func (l *lexer) Lex(yylval *asyncpiSymType) int {
	var token tok
	token, yylval.strval, _, _ = l.scanner.Scan()
	return int(token)
}

// Error handles error.
func (l *lexer) Error(err string) {
	l.Errors <- &ParseError{Err: err, Pos: l.scanner.pos}
}
