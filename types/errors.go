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

package types

import (
	"fmt"
)

// InferUnTypedError is the type of error when type inference is
// applied on an untyped Process.
type InferUntypedError struct {
	Name string
}

func (e InferUntypedError) Error() string {
	return fmt.Sprintf("infer error: name %s untyped", e.Name)
}

// TypeError is the type of error when analysing the behavioural type
// of an asyncpi Process.
type TypeError struct {
	T, U Type
	Msg  string
}

func (e TypeError) Error() string {
	return fmt.Sprintf("type error: type %s and %s does not match (%s)",
		e.T, e.U, e.Msg)
}

// TypeArityError is the type of error when process parameter arity
// does not match when unifying.
type TypeArityError struct {
	Got      int
	Expected int
	Msg      string
}

func (e TypeArityError) Error() string {
	return fmt.Sprintf("type error: arity mismatch (got=%d, expected=%d) (%s)",
		e.Got, e.Expected, e.Msg)
}
