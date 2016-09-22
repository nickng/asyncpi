package asyncpi

import "fmt"

// ErrParse is a parse error.
type ErrParse struct {
	Pos TokenPos
	Err string // Error string returned from parser.
}

func (e *ErrParse) Error() string {
	return fmt.Sprintf("Parse failed at %s: %s", e.Pos, e.Err)
}

// ErrType is a type error.
type ErrType struct {
	T, U Type
	Msg  string
}

func (e *ErrType) Error() string {
	return fmt.Sprintf("Type error: type %s and %s does not match (%s)",
		e.T, e.U, e.Msg)
}
