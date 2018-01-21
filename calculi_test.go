package asyncpi

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestNilProcessCalculi(t *testing.T) {
	const procStr = `0`
	p, err := Parse(strings.NewReader(procStr))
	if err != nil {
		t.Error(err)
	}
	if want, got := procStr, p.Calculi(); want != got {
		t.Errorf("expecting calculi to be %s but got %s", want, got)
	}
}

func TestRecvCalculi(t *testing.T) {
	const procStr = `a(b,c).0`
	p, err := Parse(strings.NewReader(procStr))
	if err != nil {
		t.Error(err)
	}
	if want, got := procStr, p.Calculi(); want != got {
		t.Errorf("expecting calculi to be %s but got %s", want, got)
	}
}

func TestRepeatCalculi(t *testing.T) {
	const procStr = `!a(b,c).0`
	p, err := Parse(strings.NewReader(procStr))
	if err != nil {
		t.Error(err)
	}
	if want, got := procStr, p.Calculi(); want != got {
		t.Errorf("expecting calculi to be %s but got %s", want, got)
	}
}

func TestRepeatParCalculi(t *testing.T) {
	const procStr = `!(a(u,v).0 | a<b,c>)`
	var buf bytes.Buffer
	p, err := Parse(io.TeeReader(strings.NewReader(procStr), &buf))
	if err != nil {
		if pe, ok := err.(*ParseError); ok {
			t.Logf("\n%s", string(pe.Pos.CaretDiag(buf.Bytes())))
		}
		t.Error(err)
	}
	if want, got := procStr, p.Calculi(); want != got {
		t.Errorf("expecting calculi to be %s but got %s", want, got)
	}
}

func TestRestrictCalculi(t *testing.T) {
	const procStr = `(new a)a(x,y).0`
	p, err := Parse(strings.NewReader(procStr))
	if err != nil {
		t.Error(err)
	}
	if want, got := procStr, p.Calculi(); want != got {
		t.Errorf("expecting calculi to be %s but got %s", want, got)
	}
}

func TestRestricParCalculi(t *testing.T) {
	const procStr = `(new a)(a(x,y).0 | a<b,c>)`
	p, err := Parse(strings.NewReader(procStr))
	if err != nil {
		t.Error(err)
	}
	if want, got := procStr, p.Calculi(); want != got {
		t.Errorf("expecting calculi to be %s but got %s", want, got)
	}
}

func TestSendCalculi(t *testing.T) {
	const procStr = `a<b,c>`
	p, err := Parse(strings.NewReader(procStr))
	if err != nil {
		t.Error(err)
	}
	if want, got := procStr, p.Calculi(); want != got {
		t.Errorf("expecting calculi to be %s but got %s", want, got)
	}
}
