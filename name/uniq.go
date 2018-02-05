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

import (
	"fmt"

	"go.nickng.io/asyncpi"
)

type setter interface {
	SetName(string)
}

func MakeNamesUnique(p asyncpi.Process) error {
	if err := Walk(new(uniqueNamer), p); err != nil {
		return err
	}
	return nil
}

type uniqueNamer struct {
	names map[asyncpi.Name]string
}

func (u *uniqueNamer) VisitName(n asyncpi.Name) error {
	if u.names == nil {
		u.names = make(map[asyncpi.Name]string)
	}
	if _, exists := u.names[n]; exists {
		// name stored and name visiting should have same Ident
		return nil
	}
	s := fmt.Sprintf("%s_%d", n.Ident(), len(u.names))
	u.names[n] = s
	if uniq, canSetName := n.(setter); canSetName {
		uniq.SetName(s)
		return nil
	}
	return asyncpi.ImmutableNameError{Name: n}
}
