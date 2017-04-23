package asyncpi

import (
	"strings"
	"testing"
)

func TestAnnotate(t *testing.T) {
	input := `a(x).x().a<y>`
	proc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	bproc := Bind(proc)
	IdentifySorts(bproc)
	AnnotateName(bproc, new(Uniquefier))
	if expect, got := 1, len(bproc.FreeNames()); expect != got {
		t.Fatalf("Expecting %d unique free names, but got %d: %s", expect, got, bproc.Calculi())
	}
	if expect, got := 1, len(bproc.FreeVars()); expect != got {
		t.Fatalf("Expecting %d unique free vars, but got %d: %s", expect, got, bproc.Calculi())
	}
}

func TestAnnotatePar(t *testing.T) {
	input := `a<x> | a<x> | x<a>`
	proc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	bproc := Bind(proc)
	IdentifySorts(bproc)
	AnnotateName(bproc, new(Uniquefier))
	if expect, got := 3, len(bproc.FreeNames()); expect != got {
		t.Fatalf("Expecting %d unique free names, but got %d: %s", expect, got, bproc.Calculi())
	}
	if expect, got := 3, len(bproc.FreeVars()); expect != got {
		t.Fatalf("Expecting %d unique free vars, but got %d: %s", expect, got, bproc.Calculi())
	}
}
