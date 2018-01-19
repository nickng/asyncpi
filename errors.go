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

// UnknownProcessTypeError is the type of error for an unknown process type.
type UnknownProcessTypeError struct {
	Caller string
	Proc   Process
}

func (e UnknownProcessTypeError) Error() string {
	return fmt.Sprintf("%s: Unknown process type: %T", e.Caller, e.Proc)
}
