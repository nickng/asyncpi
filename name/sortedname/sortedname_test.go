package sortedname

import (
	"testing"

	"go.nickng.io/asyncpi"
)

func TestDefaultSort(t *testing.T) {
	n := New(constName("name"))
	if expect, got := 1, len(asyncpi.FreeNames(n)); expect != got {
		t.Errorf("Expecting %s to have %d free names but got %d. fn(%s) = %s",
			n.Ident(), expect, got, n.Ident(), asyncpi.FreeNames(n))
	}
	if expect, got := 0, len(asyncpi.FreeVars(n)); expect != got {
		t.Errorf("Expecting %s to have %d free vars but got %d. fv(%s) = %s",
			n.Ident(), expect, got, n.Ident(), asyncpi.FreeVars(n))
	}
}

func TestNilSort(t *testing.T) {
	p := asyncpi.NewNilProcess()
	if err := InferSortsByUsage(p); err != nil {
		t.Fatalf("cannot infer sort: %v", err)
	}
	t.Logf("fn(%s) = %s", p.Calculi(), p.FreeNames())
	t.Logf("fv(%s) = %s", p.Calculi(), p.FreeVars())
	if expect, got := 0, len(p.FreeNames()); expect != got {
		t.Errorf("Expecting %s to have %d free names but got %d.",
			p.Calculi(), expect, got)
	}
	if expect, got := 0, len(p.FreeVars()); expect != got {
		t.Errorf("Expecting %s to have %d free vars but got %d.",
			p.Calculi(), expect, got)
	}
}

func TestParSort(t *testing.T) {
	pLeft, pRight := asyncpi.NewSend(constName("a")), asyncpi.NewRecv(constName("b"), asyncpi.NewNilProcess())
	pLeft.Vals = append(pLeft.Vals, constName("c"), constName("d"), constName("e"))
	pRight.Vars = append(pRight.Vars, constName("x"), constName("y"), constName("z"))
	p := asyncpi.NewPar(pLeft, pRight)
	if err := InferSortsByUsage(p); err != nil {
		t.Fatalf("cannot infer sort: %v", err)
	}
	t.Logf("fn(%s) = %s", p.Calculi(), p.FreeNames())
	t.Logf("fv(%s) = %s", p.Calculi(), p.FreeVars())
	if expect, got := 2, len(p.FreeNames()); expect != got {
		t.Errorf("Expecting %s to have %d free names but got %d.",
			p.Calculi(), expect, got)
	}
	if expect, got := 3, len(p.FreeVars()); expect != got {
		t.Errorf("Expecting %s to have %d free vars but got %d",
			p.Calculi(), expect, got)
	}
}

func TestParSortOverlap(t *testing.T) {
	pLeft, pRight := asyncpi.NewSend(constName("a")), asyncpi.NewRecv(constName("a"), asyncpi.NewNilProcess())
	pLeft.Vals = append(pLeft.Vals, constName("c"), constName("d"), constName("e"))
	pRight.Vars = append(pRight.Vars, constName("x"), constName("y"), constName("z"))
	p := asyncpi.NewPar(pLeft, pRight)
	if err := InferSortsByUsage(p); err != nil {
		t.Fatalf("cannot infer sort: %v", err)
	}
	t.Logf("fn(%s) = %s", p.Calculi(), p.FreeNames())
	t.Logf("fv(%s) = %s", p.Calculi(), p.FreeVars())
	if expect, got := 1, len(p.FreeNames()); expect != got {
		t.Errorf("Expecting %s to have %d free names but got %d.",
			p.Calculi(), expect, got)
	}
	if expect, got := 3, len(p.FreeVars()); expect != got {
		t.Errorf("Expecting %s to have %d free vars but got %d.",
			p.Calculi(), expect, got)
	}
}

func TestRestrictSort(t *testing.T) {
	p := asyncpi.NewRestrict(constName("n"), asyncpi.NewNilProcess())
	if err := InferSortsByUsage(p); err != nil {
		t.Fatalf("cannot infer sort: %v", err)
	}
	t.Logf("fn(%s) = %s", p.Calculi(), p.FreeNames())
	t.Logf("fv(%s) = %s", p.Calculi(), p.FreeVars())
	if expect, got := 0, len(p.FreeNames()); expect != got {
		t.Errorf("Expecting %s to have %d free names but got %d.",
			p.Calculi(), expect, got)
	}
	if expect, got := 0, len(p.FreeVars()); expect != got {
		t.Errorf("Expecting %s to have %d free vars but got %d",
			p.Calculi(), expect, got)
	}
}

func TestRepeatSort(t *testing.T) {
	p := asyncpi.NewRepeat(asyncpi.NewNilProcess())
	if err := InferSortsByUsage(p); err != nil {
		t.Fatalf("cannot infer sort: %v", err)
	}
	t.Logf("fn(%s) = %s", p.Calculi(), p.FreeNames())
	t.Logf("fv(%s) = %s", p.Calculi(), p.FreeVars())
	if expect, got := 0, len(p.FreeNames()); expect != got {
		t.Errorf("Expecting %s to have %d free names but got %d.",
			p.Calculi(), expect, got)
	}
	if expect, got := 0, len(p.FreeVars()); expect != got {
		t.Errorf("Expecting %s to have %d free vars but got %d",
			p.Calculi(), expect, got)
	}
}

func TestSendSort(t *testing.T) {
	p := asyncpi.NewSend(constName("u"))
	p.Vals = append(p.Vals, constName("v"))
	if err := InferSortsByUsage(p); err != nil {
		t.Fatalf("cannot infer sort: %v", err)
	}
	t.Logf("fn(%s) = %s", p.Calculi(), p.FreeNames())
	t.Logf("fv(%s) = %s", p.Calculi(), p.FreeVars())
	if expect, got := 1, len(p.FreeNames()); expect != got {
		t.Errorf("Expecting %s to have %d free names but got %d.",
			p.Calculi(), expect, got)
	}
	if p.FreeNames()[0] != p.Chan {
		t.Errorf("Expecting fn(%s) to be %s but got %s",
			p.Calculi(), p.Chan, p.FreeNames()[0])
	}
	if expect, got := 1, len(p.FreeVars()); expect != got {
		t.Errorf("Expecting %s to have %d free vars but got %d",
			p.Calculi(), expect, got)
	}
	if p.FreeVars()[0] != p.Vals[0] {
		t.Errorf("Expecting fv(%s) to be %s but got %s",
			p.Calculi(), p.Vals[0], p.FreeVars()[0])
	}
}

func TestRecvSort(t *testing.T) {
	p := asyncpi.NewRecv(constName("u"), asyncpi.NewNilProcess())
	p.Vars = append(p.Vars, constName("x"))
	if err := InferSortsByUsage(p); err != nil {
		t.Fatalf("cannot infer sort: %v", err)
	}
	t.Logf("fn(%s) = %s", p.Calculi(), p.FreeNames())
	t.Logf("fv(%s) = %s", p.Calculi(), p.FreeVars())
	if expect, got := 1, len(p.FreeNames()); expect != got {
		t.Errorf("Expecting %s to have %d free names but got %d.",
			p.Calculi(), expect, got)
	}
	if p.FreeNames()[0] != p.Chan {
		t.Errorf("Expecting fn(%s) to be %s but got %s",
			p.Calculi(), p.Chan, p.FreeNames()[0])
	}
	if expect, got := 0, len(p.FreeVars()); expect != got {
		t.Errorf("Expecting %s to have %d free vars but got %d",
			p.Calculi(), expect, got)
	}
}
