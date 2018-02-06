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
	"bytes"
	"fmt"

	"go.nickng.io/asyncpi"
)

// ProcType returns the process type of the Process p.
func ProcType(p asyncpi.Process) (string, error) {
	switch p := p.(type) {
	case *asyncpi.NilProcess:
		return "0", nil
	case *asyncpi.Send:
		return fmt.Sprintf("%s!%s", p.Chan.Ident(), p.Chan.(TypedName).Type()), nil
	case *asyncpi.Recv:
		proc, err := ProcType(p.Cont)
		if err != nil {
			return proc, err
		}
		return fmt.Sprintf("%s?%s; %s", p.Chan.Ident(), p.Chan.(TypedName).Type(), proc), nil
	case *asyncpi.Par:
		var buf bytes.Buffer
		for i, ps := range p.Procs {
			if i != 0 {
				buf.WriteRune('|')
			}
			proc, err := ProcType(ps)
			if err != nil {
				return proc, err
			}
			buf.WriteString(proc)
		}
		return buf.String(), nil
	case *asyncpi.Repeat:
		proc, err := ProcType(p.Proc)
		if err != nil {
			return proc, err
		}
		return "*" + proc, nil
	case *asyncpi.Restrict:
		proc, err := ProcType(p.Proc)
		if err != nil {
			return proc, err
		}
		return fmt.Sprintf("(ν%s:%s) %s", p.Name.Ident(), p.Name.(TypedName).Type(), proc), nil
	default:
		return "", asyncpi.UnknownProcessError{Proc: p}
	}
}
