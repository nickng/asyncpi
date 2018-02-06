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

package name

import "go.nickng.io/asyncpi"

// A Visitor's VisitName method is is invoked for each Name encounted by Walk.
type Visitor interface {
	VisitName(n asyncpi.Name) error
}

// Walk traverses a Process p in breadth-first order,
// and applies v.VisitName(n) on each Name n encountered.
func Walk(v Visitor, proc asyncpi.Process) error {
	procs := []asyncpi.Process{proc}
	for len(procs) > 0 {
		proc, procs = procs[0], procs[1:]
		switch p := proc.(type) {
		case *asyncpi.NilProcess:
			// finish
		case *asyncpi.Repeat:
			procs = append(procs, p.Proc)
		case *asyncpi.Par:
			procs = append(procs, p.Procs...)
		case *asyncpi.Recv:
			if err := v.VisitName(p.Chan); err != nil {
				return err
			}
			for i := range p.Vars {
				if err := v.VisitName(p.Vars[i]); err != nil {
					return err
				}
			}
			procs = append(procs, p.Cont)
		case *asyncpi.Send:
			if err := v.VisitName(p.Chan); err != nil {
				return err
			}
			for i := range p.Vals {
				if err := v.VisitName(p.Vals[i]); err != nil {
					return err
				}
			}
		case *asyncpi.Restrict:
			if err := v.VisitName(p.Name); err != nil {
				return err
			}
		default:
			return asyncpi.InvalidProcTypeError{Caller: "name.Walk", Proc: p}
		}
	}
	return nil
}
