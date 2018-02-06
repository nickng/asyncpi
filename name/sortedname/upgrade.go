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

package sortedname

import (
	"go.nickng.io/asyncpi"
)

// Upgrade wraps all Names in the Process p into SortedNames.
func upgrade(p asyncpi.Process) error {
	procs := []asyncpi.Process{p}
	for len(procs) > 0 {
		p, procs = procs[0], procs[1:]
		switch p := p.(type) {
		case *asyncpi.NilProcess:
			// nothing to do
		case *asyncpi.Repeat:
			procs = append(procs, p.Proc)
		case *asyncpi.Par:
			procs = append(procs, p.Procs...)
		case *asyncpi.Recv:
			p.Chan = New(p.Chan)
			var vars []asyncpi.Name
			for i := range p.Vars {
				vars = append(vars, New(p.Vars[i]))
			}
			p.Vars = vars
			procs = append(procs, p.Cont)
		case *asyncpi.Send:
			p.Chan = New(p.Chan)
			var vals []asyncpi.Name
			for i := range p.Vals {
				vals = append(vals, New(p.Vals[i]))
			}
			p.Vals = vals
		case *asyncpi.Restrict:
			p.Name = New(p.Name)
			procs = append(procs, p.Proc)
		default:
			return asyncpi.InvalidProcTypeError{Caller: "sortedname.upgrade", Proc: p}
		}
	}
	return nil
}
