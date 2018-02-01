package asyncpi

import (
	"strings"
	"testing"
)

// Tests reduction of (send | recv)
func TestReduceSendRecv(t *testing.T) {
	const proc = `(new a)(a<b> | a(x).x().0)`
	p, err := Parse(strings.NewReader(proc))
	if err != nil {
		t.Fatal(err)
	}
	if changed, err := reduceOnce(p); err != nil {
		t.Fatalf("cannot reduce: %v", err)
	} else if !changed {
		t.Fatalf("expects %s to reduce but unchanged", p.Calculi())
	}
	p, err = SimplifyBySC(p)
	if err != nil {
		t.Fatalf("cannot simplify process: %v", err)
	}
	t.Logf("%s reduces to %s", proc, p.Calculi())
	p2, ok := p.(*Recv)
	if !ok {
		t.Fatalf("expects *Recv but got %s (type %T)", p.Calculi(), p)
	}
	if want, got := "b", p2.Chan.Name(); want != got {
		t.Fatalf("expects receive on %s but got %s", want, got)
	}
	_, ok = p2.Cont.(*NilProcess)
	if !ok {
		t.Fatalf("expects *NilProcess but got %s (type %T)", p2.Cont.Calculi(), p2.Cont)
	}
}

// Test reduction of (recv | send)
func TestReduceRecvSend(t *testing.T) {
	const proc = `(new a)(a(x).x<> | a<b>)`
	p, err := Parse(strings.NewReader(proc))
	if err != nil {
		t.Fatal(err)
	}
	if changed, err := reduceOnce(p); err != nil {
		t.Fatalf("cannot reduce: %v", err)
	} else if !changed {
		t.Fatalf("expects %s to reduce but unchanged", p.Calculi())
	}
	p, err = SimplifyBySC(p)
	if err != nil {
		t.Fatalf("cannot simplify process: %v", err)
	}
	t.Logf("%s reduces to %s", proc, p.Calculi())
	p2, ok := p.(*Send)
	if !ok {
		t.Fatalf("expects *Send but got %s (type %T)", p.Calculi(), p)
	}
	if want, got := "b", p2.Chan.Name(); want != got {
		t.Fatalf("expects send on %s but got %s", want, got)
	}
}

// Test reduction (and simplify) where all names are bound.
func TestReduceBoundRecvSend(t *testing.T) {
	const proc = `(new a)(new b)(a(x).x<> | a<b>)`
	p, err := Parse(strings.NewReader(proc))
	if err != nil {
		t.Fatal(err)
	}
	if changed, err := reduceOnce(p); err != nil {
		t.Fatalf("cannot reduce: %v", err)
	} else if !changed {
		t.Fatalf("expects %s to reduce but unchanged", p.Calculi())
	}
	p, err = SimplifyBySC(p)
	if err != nil {
		t.Fatalf("cannot simplify process: %v", err)
	}
	t.Logf("%s reduces to %s", proc, p.Calculi())
	p2, ok := p.(*Restrict)
	if !ok {
		t.Fatalf("expects a *Restrict but got %s (type %T)", p.Calculi(), p)
	}
	if want, got := "b", p2.Name.Name(); want != got {
		t.Fatalf("expects restrict on %s but got %s", want, got)
	}
	p3, ok := p2.Proc.(*Send)
	if !ok {
		t.Fatalf("expects *Send but got %s (type %T)", p2.Calculi(), p2)
	}
	if want, got := "b", p3.Chan.Name(); want != got {
		t.Fatalf("expects send on %s but got %s", want, got)
	}
}

func TestReduceFreeSendRecv(t *testing.T) {
	const proc = `a<b> | a(x).0`
	p, err := Parse(strings.NewReader(proc))
	if err != nil {
		t.Fatal(err)
	}
	if changed, err := reduceOnce(p); err != nil {
		t.Fatalf("cannot reduce: %v", err)
	} else if !changed {
		t.Fatalf("expects %s to reduce but unchanged", p.Calculi())
	}
	t.Logf("%s reduces to %s", proc, p.Calculi())
}

func TestReduceNone(t *testing.T) {
	const proc = `(new a)(a<b> | b(x).0)`
	p, err := Parse(strings.NewReader(proc))
	if err != nil {
		t.Fatal(err)
	}
	if changed, err := reduceOnce(p); err != nil {
		t.Fatalf("cannot reduce: %v", err)
	} else if changed {
		t.Fatalf("expects %s to not reduce but reduced to %s", proc, p.Calculi())
	}
	t.Logf("%s reduces to %s (no change)", proc, p.Calculi())
}

func TestReduceBoundRecvRecv(t *testing.T) {
	const proc = `(new a)(a(z).0 | a(x).0)`
	p, err := Parse(strings.NewReader(proc))
	if err != nil {
		t.Fatal(err)
	}
	if changed, err := reduceOnce(p); err != nil {
		t.Fatalf("cannot reduce: %v", err)
	} else if changed {
		t.Fatalf("expects %s to not reduce but reduced to %s", proc, p.Calculi())
	}
	t.Logf("%s reduces to %s (no change)", proc, p.Calculi())
}

// Non-shared name cannot reduce.
func TestReduceFreeRecvRecv(t *testing.T) {
	const proc = `a(z).0 | a(x).0`
	p, err := Parse(strings.NewReader(proc))
	if err != nil {
		t.Fatal(err)
	}
	if changed, err := reduceOnce(p); err != nil {
		t.Fatalf("cannot reduce: %v", err)
	} else if changed {
		t.Fatalf("expects %s to note reduce but it changed to %s", proc, p.Calculi())
	}
	if changed, err := reduceOnce(p); err != nil {
		t.Fatalf("cannot reduce: %v", err)
	} else if changed {
		t.Fatalf("expects %s to not reduce but reduced to %s", proc, p.Calculi())
	}
	t.Logf("%s reduces to %s (no change)", proc, p.Calculi())
}

func TestReduceMultiple(t *testing.T) {
	const proc = `(new a)(a<b,c>|a(x,y).x(z).z<y> | b<d> | d(z).0)`
	p, err := Parse(strings.NewReader(proc))
	if err != nil {
		t.Fatal(err)
	}
	if changed, err := reduceOnce(p); err != nil {
		t.Fatalf("cannot reduce: %v", err)
	} else if !changed {
		t.Fatalf("expects %s to reduce but unchanged", p.Calculi())
	}
	p, err = SimplifyBySC(p)
	if err != nil {
		t.Fatalf("cannot simplify process: %v", err)
	}
	t.Logf("%s reduces to %s", proc, p.Calculi())
	procPrev := p.Calculi()
	if changed, err := reduceOnce(p); err != nil {
		t.Fatalf("cannot reduce: %v", err)
	} else if !changed {
		t.Fatalf("expects %s to reduce but unchanged", p.Calculi())
	}
	p, err = SimplifyBySC(p)
	if err != nil {
		t.Fatalf("cannot simplify process: %v", err)
	}
	t.Logf("%s reduces to %s", procPrev, p.Calculi())
	procPrev = p.Calculi()
	if changed, err := reduceOnce(p); err != nil {
		t.Fatalf("cannot reduce: %v", err)
	} else if !changed {
		t.Fatalf("expects %s to reduce but unchanged", p.Calculi())
	}
	p, err = SimplifyBySC(p)
	if err != nil {
		t.Fatalf("cannot simplify process: %v", err)
	}
	t.Logf("%s reduces to %s", procPrev, p.Calculi())
	if _, ok := p.(*NilProcess); !ok {
		t.Fatalf("expects *NilProcess but got %s (type %T)", p.Calculi(), p)
	}
}
