// Copyright 2018 Nicholas Ng <nickng@nickng.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package errors provides internal error handlign routines.
package errors

import (
	"fmt"
	"io"
)

// propagatedError is an error wrapper from github.com/pkg/errors.
type propagatedError struct {
	cause error
	msg   string
}

func (e *propagatedError) Error() string {
	return e.msg + ": " + e.cause.Error()
}

func (e *propagatedError) Cause() error {
	return e.cause
}

func (e *propagatedError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", e.Cause())
			io.WriteString(s, e.msg)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, e.Error())
	}
}

// Wrap is function to propagate error to upper levels.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &propagatedError{
		cause: err,
		msg:   msg,
	}
}
