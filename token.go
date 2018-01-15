package asyncpi

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Tokens for use with lexer and parser.

// tok is a lexical token.
type tok int

const (
	// kILLEGAL is a special token for errors.
	kILLEGAL tok = iota
)

var eof = rune(0)

// TokenPos is a pair of coordinate to identify start of token.
type TokenPos struct {
	Char  int
	Lines []int
}

// CaretDiag returns the input b with caret to locate error position for diagnosis.
func (pos TokenPos) CaretDiag(b []byte) []byte {
	var lastLine bytes.Buffer

	for _, l := range pos.Lines {
		if l > 0 {
			lastLine.Reset() // New line will replace last line.
		}
		for c := 0; c < l; {
			r, size := utf8.DecodeRune(b)
			if r != '\n' {
				lastLine.WriteRune(r)
			}
			b = b[size:]
			c += size
		}
		b = b[1:] // newline
	}

	var errbuf bytes.Buffer
	var caret bytes.Buffer
	column := 0
LINE:
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if r == '\n' {
			if column == 0 {
			} else {
				break LINE
			}
		} else {
			errbuf.WriteRune(r)
		}
		if column == pos.Char-1 {
			caret.WriteRune('↑')
		} else if column < pos.Char-1 {
			caret.WriteRune(' ')
		}
		b = b[size:]
		column += size
	}

	var diag bytes.Buffer
	prefix := strings.Repeat(" ", len(pos.String()))
	if lastLine.String() != "" {
		diag.WriteString(fmt.Sprintf("%s   %s\n", prefix, lastLine.String()))
	}
	diag.WriteString(fmt.Sprintf("%s → %s\n", pos.String(), errbuf.String()))
	diag.WriteString(fmt.Sprintf("%s   %s\n", prefix, caret.String()))
	return diag.Bytes()
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
