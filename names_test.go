package asyncpi

import (
	"strings"
	"testing"
)

func TestUpdateName(t *testing.T) {
	input := `a(x).x().a<y>`
	proc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	bproc := Bind(proc)
	InferSorts(bproc)
	if err := UpdateName(bproc, new(Uniquefier)); err != nil {
		t.Fatalf("cannot update name: %v", err)
	}
	if expect, got := 1, len(bproc.FreeNames()); expect != got {
		t.Fatalf("Expecting %d unique free names, but got %d: %s", expect, got, bproc.Calculi())
	}
	if expect, got := 1, len(bproc.FreeVars()); expect != got {
		t.Fatalf("Expecting %d unique free vars, but got %d: %s", expect, got, bproc.Calculi())
	}
}

func TestUpdateNamePar(t *testing.T) {
	input := `a<x> | a<x> | x<a>`
	proc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	bproc := Bind(proc)
	if err := InferSorts(bproc); err != nil {
		t.Fatalf("cannot infer sort: %v", err)
	}
	if err := UpdateName(bproc, new(Uniquefier)); err != nil {
		t.Fatalf("cannot update name: %v", err)
	}
	if expect, got := 3, len(bproc.FreeNames()); expect != got {
		t.Fatalf("Expecting %d unique free names, but got %d: %s", expect, got, bproc.Calculi())
	}
	if expect, got := 3, len(bproc.FreeVars()); expect != got {
		t.Fatalf("Expecting %d unique free vars, but got %d: %s", expect, got, bproc.Calculi())
	}
}
