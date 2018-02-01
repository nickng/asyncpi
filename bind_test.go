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
	bp0 := Bind(p)
	p1 := bp0.Calculi()
	bp1 := Bind(bp0)
	p2 := bp1.Calculi()
	if p0 != p1 {
		t.Errorf("expect Bind to be idempotent but got: \nBefore:\t%s\nAfter:\t%s", p0, p1)
	}
	if p1 != p2 {
		t.Errorf("expect Bind to be idempotent but got: \nBefore:\t%s\nAfter:\t%s", p1, p2)
	}
}
