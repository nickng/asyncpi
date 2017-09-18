package asyncpi

import "fmt"

// ParseError is the type of error when parsing an asyncpi process.
type ParseError struct {
	Pos TokenPos
	Err string // Error string returned from parser.
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Parse failed at %s: %s", e.Pos, e.Err)
}

// TypeError is the type of error when analysing the behavioural type of an
// asyncpi process.
type TypeError struct {
	T, U Type
	Msg  string
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("Type error: type %s and %s does not match (%s)",
		e.T, e.U, e.Msg)
}

// TypeArityError is the type of error when process parameter arity does not
// match when unifying.
type TypeArityError struct {
	Got, Expected int
	Msg           string
}

func (e *TypeArityError) Error() string {
	return fmt.Sprintf("Type error: type arity mismatch (got=%d, expected=%d) (%s)",
		e.Got, e.Expected, e.Msg)
}

// UnknownProcessTypeError is the type of error for an unknown process type.
type UnknownProcessTypeError struct {
	Caller string
	Proc   Process
}

func (e UnknownProcessTypeError) Error() string {
	return fmt.Sprintf("%s: Unknown process type: %T", e.Caller, e.Proc)
}
