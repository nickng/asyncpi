package asyncpi

import "testing"

func TestDefaultSort(t *testing.T) {
	n := newPiName("name")
	if expect, got := 1, len(FreeNames(n)); expect != got {
		t.Errorf("Expecting %s to have %d free names but got %d. fn(%s) = %s",
			n.Ident(), expect, got, n.Ident(), FreeNames(n))
	}
	if expect, got := 0, len(FreeVars(n)); expect != got {
		t.Errorf("Expecting %s to have %d free vars but got %d. fv(%s) = %s",
			n.Ident(), expect, got, n.Ident(), FreeVars(n))
	}
}

func TestNilSort(t *testing.T) {
	p := NewNilProcess()
	IdentifySorts(p)
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
	pLeft, pRight := NewSend(newPiName("a")), NewRecv(newPiName("b"), NewNilProcess())
	pLeft.Vals = append(pLeft.Vals, newPiName("c"), newPiName("d"), newPiName("e"))
	pRight.Vars = append(pRight.Vars, newPiName("x"), newPiName("y"), newPiName("z"))
	p := NewPar(pLeft, pRight)
	IdentifySorts(p)
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
	pLeft, pRight := NewSend(newPiName("a")), NewRecv(newPiName("a"), NewNilProcess())
	pLeft.Vals = append(pLeft.Vals, newPiName("c"), newPiName("d"), newPiName("e"))
	pRight.Vars = append(pRight.Vars, newPiName("x"), newPiName("y"), newPiName("z"))
	p := NewPar(pLeft, pRight)
	IdentifySorts(p)
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
	p := NewRestrict(newPiName("n"), NewNilProcess())
	IdentifySorts(p)
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
	p := NewRepeat(NewNilProcess())
	IdentifySorts(p)
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
	p := NewSend(newPiName("u"))
	p.Vals = append(p.Vals, newPiName("v"))
	IdentifySorts(p)
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
	p := NewRecv(newPiName("u"), NewNilProcess())
	p.Vars = append(p.Vars, newPiName("x"))
	IdentifySorts(p)
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

func TestNameVarSort(t *testing.T) {
	a := newPiName("a")
	b := newPiName("b")
	x := newPiName("x")
	y := newPiName("y")
	z := newPiName("z")
	p := NewRecv(a, NewNilProcess())
	p.Vars = append(p.Vars, b, x, y, z)
	UpdateName(p, new(NameVarSorter))
	if expect, got := nameSort, a.s; expect != got {
		t.Errorf("Expecting %s sort to be %d but got %d.", a.Ident(), expect, got)
	}
	if expect, got := nameSort, b.s; expect != got {
		t.Errorf("Expecting %s sort to be %d but got %d.", b.Ident(), expect, got)
	}
	if expect, got := varSort, x.s; expect != got {
		t.Errorf("Expecting %s sort to be %d but got %d.", x.Ident(), expect, got)
	}
	if expect, got := varSort, y.s; expect != got {
		t.Errorf("Expecting %s sort to be %d but got %d.", y.Ident(), expect, got)
	}
	if expect, got := varSort, z.s; expect != got {
		t.Errorf("Expecting %s sort to be %d but got %d.", z.Ident(), expect, got)
	}
}
