package asyncpi

import (
	"strings"
	"testing"
)

func TestBindIdempotent(t *testing.T) {
	const proc = `(new a)!(a<>|a().0)`
	p, err := Parse(strings.NewReader(proc))
	if err != nil {
		t.Fatal(err)
	}
	p0 := p.Calculi()
	if err := Bind(&p); err != nil {
		t.Fatal(err)
	}
	p1 := p.Calculi()
	if err := Bind(&p); err != nil {
		t.Fatal(err)
	}
	p2 := p.Calculi()
	if p0 != p1 {
		t.Errorf("expect Bind to be idempotent but got: \nBefore:\t%s\nAfter:\t%s", p0, p1)
	}
	if p1 != p2 {
		t.Errorf("expect Bind to be idempotent but got: \nBefore:\t%s\nAfter:\t%s", p1, p2)
	}
}

func TestBindRebind(t *testing.T) {
	const proc = `(new a)(a<> | a().(new a)a().0)`
	p, err := Parse(strings.NewReader(proc))
	if err != nil {
		t.Fatal(err)
	}
	if err := Bind(&p); err != nil {
		t.Fatalf("cannot bind: %v", err)
	}
	type setter interface {
		SetName(string)
	}
	if s, ok := p.(*Restrict).Name.(setter); ok {
		s.SetName("b")
	}
	if want, got := "b", p.(*Restrict).Proc.(*Par).Procs[1].(*Recv).Chan.Ident(); want != got {
		t.Fatalf("expects %s but got %s for %s", want, got, p.(*Restrict).Proc.(*Par).Procs[1].(*Recv).Calculi())
	}
	if want, got := "a", p.(*Restrict).Proc.(*Par).Procs[1].(*Recv).Cont.(*Restrict).Name.Ident(); want != got {
		t.Fatalf("expects %s but got %s for %s", want, got, p.(*Restrict).Proc.(*Par).Procs[1].(*Recv).Cont.(*Restrict).Calculi())
	}
}
