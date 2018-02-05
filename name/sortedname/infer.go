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
	"strings"
	"unicode/utf8"

	"go.nickng.io/asyncpi"
	"go.nickng.io/asyncpi/name"
)

// InferSortsByUsage puts names in a Process into their respective sort {name,var}.
func InferSortsByUsage(p asyncpi.Process) error {
	if err := upgrade(p); err != nil {
		return err
	}
	nameVar := make(map[asyncpi.Name]bool)
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
			for i := range p.Vars {
				nameVar[p.Vars[i]] = true
				if s, canSetSort := p.Vars[i].(setter); canSetSort {
					s.SetSort(VarSort)
				} else {
					return asyncpi.ImmutableNameError{Name: p.Vars[i]}
				}
			}
			procs = append(procs, p.Cont)
		case *asyncpi.Send:
			for i := range p.Vals {
				if _, ok := nameVar[p.Vals[i]]; !ok {
					nameVar[p.Vals[i]] = true
					if s, canSetSort := p.Vals[i].(setter); canSetSort {
						s.SetSort(VarSort)
					} else {
						return asyncpi.ImmutableNameError{Name: p.Vals[i]}
					}
				}
			}
		case *asyncpi.Restrict:
			nameVar[p.Name] = false // new name = not var
			procs = append(procs, p.Proc)
		default:
			return asyncpi.UnknownProcessTypeError{Caller: "sortedname.InferSortsByUsage", Proc: p}
		}
	}
	return nil
}

func InferSortsByPrefix(p asyncpi.Process) error {
	if err := name.Walk(byPrefix{}, p); err != nil {
		return err
	}
	return nil
}

// byPrefix is a name.Visitor implementation which puts names in sorts.
// A Name is a name/var depending on its prefix:
//   names={a,b,c,...} vars={...,x,y,z}
type byPrefix struct{}

func (v byPrefix) VisitName(n asyncpi.Name) error {
	r, _ := utf8.DecodeRuneInString(n.Ident())
	if strings.ContainsRune("nopqrstuvwxyz", r) {
		if s, canSetSort := n.(setter); canSetSort {
			s.SetSort(VarSort)
			return nil
		}
	} else {
		if s, canSetSort := n.(setter); canSetSort {
			s.SetSort(NameSort)
			return nil
		}
	}
	return asyncpi.ImmutableNameError{Name: n}
}
