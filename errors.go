package asyncpi

import (
	"errors"
	"fmt"
)

// ParseError is the type of error when parsing an asyncpi process.
type ParseError struct {
	Pos TokenPos
	Err string // Error string returned from parser.
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Parse failed at %s: %s", e.Pos, e.Err)
}

// ImmutableNameError is the type of error when trying
// to modify a name without setter methods.
type ImmutableNameError struct {
	Name Name
}

func (e ImmutableNameError) Error() string {
	return fmt.Sprintf("cannot modify name %v: immutable implementation of Name", e.Name.Ident())
}

// InvalidProcTypeError is the type of error for an unknown process type.
type InvalidProcTypeError struct {
	Caller string
	Proc   Process
}

func (e InvalidProcTypeError) Error() string {
	return fmt.Sprintf("%s: Unknown process type: %T", e.Caller, e.Proc)
}

var ErrInvalid = errors.New("invalid argument")
