package asyncpi

import (
	"bufio"
	"bytes"
	"io"
)

// scanner is a lexical scanner.
type scanner struct {
	r   *bufio.Reader
	pos TokenPos
}

// newScanner returns a new instance of Scanner.
func newScanner(r io.Reader) *scanner {
	return &scanner{r: bufio.NewReader(r), pos: TokenPos{Char: 0, Lines: []int{}}}
}

// read reads the next rune from the buffered reader.
// Returns the rune(0) if reached the end or error occurs.
func (s *scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	if ch == '\n' {
		s.pos.Lines = append(s.pos.Lines, s.pos.Char)
		s.pos.Char = 0
	} else {
		s.pos.Char++
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *scanner) unread() {
	_ = s.r.UnreadRune()
	if s.pos.Char == 0 {
		s.pos.Char = s.pos.Lines[len(s.pos.Lines)-1]
		s.pos.Lines = s.pos.Lines[:len(s.pos.Lines)-1]
	} else {
		s.pos.Char--
	}
}

// Scan returns the next token and parsed value.
func (s *scanner) Scan() (token tok, value string, startPos, endPos TokenPos) {
	ch := s.read()

	if isWhitespace(ch) {
		s.skipWhitespace()
		ch = s.read()
	}
	if isAlphaNum(ch) {
		s.unread()
		return s.scanName()
	}

	// Track token positions.
	startPos = s.pos
	defer func() { endPos = s.pos }()

	switch ch {
	case eof:
		return 0, "", startPos, endPos
	case '<':
		return kLANGLE, string(ch), startPos, endPos
	case '>':
		return kRANGLE, string(ch), startPos, endPos
	case '(':
		return kLPAREN, string(ch), startPos, endPos
	case ')':
		return kRPAREN, string(ch), startPos, endPos
	case '.':
		return kPREFIX, string(ch), startPos, endPos
	case ';':
		return kSEMICOLON, string(ch), startPos, endPos
	case ':':
		return kCOLON, string(ch), startPos, endPos
	case '|':
		return kPAR, string(ch), startPos, endPos
	case '!':
		return kREPEAT, string(ch), startPos, endPos
	case ',':
		return kCOMMA, string(ch), startPos, endPos
	}

	return kILLEGAL, string(ch), startPos, endPos
}

func (s *scanner) scanName() (token tok, value string, startPos, endPos TokenPos) {
	var buf bytes.Buffer
	startPos = s.pos
	defer func() { endPos = s.pos }()
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isAlphaNum(ch) && !isNameSymbols(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	switch buf.String() {
	case "0":
		return kNIL, buf.String(), startPos, endPos
	case "new":
		return kNEW, buf.String(), startPos, endPos
	}
	return kNAME, buf.String(), startPos, endPos
}

func (s *scanner) skipWhitespace() {
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		}
	}
}
