package sortedname

import (
	"testing"

	"go.nickng.io/asyncpi"
)

// constName is a generic Name for testing.
type constName string

func (n constName) Ident() string {
	return string(n)
}
func (n constName) String() string {
	return string(n)
}

func TestInferSortsByPrefix(t *testing.T) {
	a := New(constName("a"))
	b := New(constName("b"))
	x := New(constName("x"))
	y := New(constName("y"))
	z := New(constName("z"))
	p := asyncpi.NewRecv(a, asyncpi.NewNilProcess())
	p.Vars = append(p.Vars, b, x, y, z)
	if err := InferSortsByPrefix(p); err != nil {
		t.Fatal(err)
	}
	if expect, got := NameSort, a.Sort(); expect != got {
		t.Errorf("Expecting %s sort to be %d but got %d.", a.Ident(), expect, got)
	}
	if expect, got := NameSort, b.Sort(); expect != got {
		t.Errorf("Expecting %s sort to be %d but got %d.", b.Ident(), expect, got)
	}
	if expect, got := VarSort, x.Sort(); expect != got {
		t.Errorf("Expecting %s sort to be %d but got %d.", x.Ident(), expect, got)
	}
	if expect, got := VarSort, y.Sort(); expect != got {
		t.Errorf("Expecting %s sort to be %d but got %d.", y.Ident(), expect, got)
	}
	if expect, got := VarSort, z.Sort(); expect != got {
		t.Errorf("Expecting %s sort to be %d but got %d.", z.Ident(), expect, got)
	}
}
