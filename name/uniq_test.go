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
	"strings"
	"testing"

	"go.nickng.io/asyncpi"
)

func TestUpdateName(t *testing.T) {
	input := `a(x).x().a<y>`
	proc, err := asyncpi.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	bproc := asyncpi.Bind(proc)
	if err := MakeNamesUnique(bproc); err != nil {
		t.Fatalf("cannot update name: %v", err)
	}
	if expect, got := 2, len(bproc.FreeNames()); expect != got {
		t.Fatalf("Expecting %d unique free names, but got %d: %s", expect, got, bproc.Calculi())
	}
	if expect, got := 0, len(bproc.FreeVars()); expect != got {
		t.Fatalf("Expecting %d unique free vars, but got %d: %s", expect, got, bproc.Calculi())
	}
	t.Logf("%s has unique names %q", bproc.Calculi(), bproc.FreeNames())
}

func TestUpdateNamePar(t *testing.T) {
	input := `a<x> | a<x> | x<a>`
	proc, err := asyncpi.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	bproc := asyncpi.Bind(proc)
	if err := MakeNamesUnique(bproc); err != nil {
		t.Fatalf("cannot update name: %v", err)
	}
	if expect, got := 6, len(bproc.FreeNames()); expect != got {
		t.Fatalf("Expecting %d unique free names, but got %d: %s", expect, got, bproc.Calculi())
	}
	if expect, got := 0, len(bproc.FreeVars()); expect != got {
		t.Fatalf("Expecting %d unique free vars, but got %d: %s", expect, got, bproc.Calculi())
	}
}
