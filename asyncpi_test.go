package asyncpi

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"go.nickng.io/asyncpi/internal/name"
)

type TestCase struct {
	Input     string
	Output    string
	FreeNames []Name
}

var TestCases map[string]TestCase

func init() {
	TestCases = map[string]TestCase{
		"NilProcess": {
			Input:     `    0 `,
			Output:    `inact`,
			FreeNames: []Name{},
		},
		"Par": {
			Input:     `b().a().0 | b<> | (new x)x(a,b,c).0`,
			Output:    `par[ par[ recv(b,[]).recv(a,[]).inact | send(b,[]) ] | restrict(x,recv(x,[a b c]).inact) ]`,
			FreeNames: newNames("a", "b"),
		},
		"Recv": {
			Input:     `a(b, c,d__).   0 `,
			Output:    `recv(a,[b c d__]).inact`,
			FreeNames: newNames("a"),
		},
		"Rep": {
			Input:     `! a().0`,
			Output:    `repeat(recv(a,[]).inact)`,
			FreeNames: []Name{name.New("a")},
		},
		"Res": {
			Input:     `(new x)  x().0 `,
			Output:    `restrict(x,recv(x,[]).inact)`,
			FreeNames: []Name{},
		},
		"Send": {
			Input:     `a<b, e_, b> `,
			Output:    `send(a,[b e_ b])`,
			FreeNames: []Name{name.New("a"), name.New("b"), name.New("e_")},
		},
	}
}

// Tests fn(a) is a
func TestFreeName(t *testing.T) {
	n := name.New("a")
	freeNames := FreeNames(n)
	if len(freeNames) == 1 && freeNames[0].Ident() != n.Ident() {
		t.Errorf("FreeName: fn(a) does not match a: `%s` vs `%s`", freeNames, n)
	}
}

// Tests fn(a) U fn(b) is a U b
func TestFreeNames(t *testing.T) {
	piNames := []Name{name.New("a"), name.New("c"), name.New("b")}
	freeNames := []Name{}
	for _, name := range piNames {
		freeNames = append(freeNames, FreeNames(name)...)
	}
	sort.Slice(freeNames, names(freeNames).Less)
	sort.Slice(piNames, names(piNames).Less)
	if len(piNames) != len(freeNames) {
		t.Errorf("FreeNames: fn(a...) and a... have different sizes")
		t.Fail()
	}
	for i := range piNames {
		if piNames[i].Ident() != freeNames[i].Ident() {
			t.Errorf("FreeNames: fn(a...) does not match a...: `%s` vs `%s`", freeNames, piNames)
		}
	}
}

// Tests parsing of nil process.
func TestParseNilProcess(t *testing.T) {
	test := TestCases["NilProcess"]
	proc, err := Parse(strings.NewReader(test.Input))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(proc.String()) != test.Output {
		t.Errorf("Parse: `%s` not parsed as nil process: `%s`\nparsed: %s",
			test.Input, test.Output, proc)
	}
}

// Tests parsing of parallel composition.
func TestParsePar(t *testing.T) {
	test := TestCases["Par"]
	proc, err := Parse(strings.NewReader(test.Input))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(proc.String()) != test.Output {
		t.Errorf("Parse: `%s` not parsed as par: `%s`\nparsed: %s",
			test.Input, test.Output, proc)
	}
}

// Tests FreeVar calculation of parallel composition.
func TestParFreeVar(t *testing.T) {
	test := TestCases["Par"]
	proc, err := Parse(strings.NewReader(test.Input))
	if err != nil {
		t.Fatal(err)
	}
	if len(proc.FreeNames()) != len(test.FreeNames) {
		t.Errorf("FreeNames(par): parsed and test case have different sizes: `%s` vs `%s`", proc.FreeNames(), test.FreeNames)
		t.Fail()
	}
	for i := range test.FreeNames {
		fn := proc.FreeNames()[i]
		if fn.Ident() != test.FreeNames[i].Ident() {
			t.Errorf("FreeNames(par): parsed and test case do not match: `%s` vs `%s`",
				fn, test.FreeNames[i])
		}
	}
}

// Tests parsing of receive.
func TestParseRecv(t *testing.T) {
	test := TestCases["Recv"]
	proc, err := Parse(strings.NewReader(test.Input))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(proc.String()) != test.Output {
		t.Errorf("Parse: `%s` not parsed as receive: `%s`\nparsed: %s",
			test.Input, test.Output, proc)
	}
}

func TestParsedRepeat(t *testing.T) {
	test := TestCases["Rep"]
	proc, err := Parse(strings.NewReader(test.Input))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(proc.String()) != test.Output {
		t.Errorf("Parse: `%s` not parsed as repeat: `%s`\nparsed: %s",
			test.Input, test.Output, proc)
	}
}

// Tests parsing of restrict.
func TestParseRestrict(t *testing.T) {
	test := TestCases["Res"]
	proc, err := Parse(strings.NewReader(test.Input))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(proc.String()) != test.Output {
		t.Errorf("Parse: `%s` not parsed as restrict: `%s`\nparsed: %s",
			test.Input, test.Output, proc)
	}
}

// Tests parsing of send.
func TestParseSend(t *testing.T) {
	test := TestCases["Send"]
	proc, err := Parse(strings.NewReader(test.Input))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(proc.String()) != test.Output {
		t.Errorf("Parse: `%s` not parsed as send `%s`.\nparsed: %s",
			test.Input, test.Output, proc)
	}
}

// Tests syntax error.
func TestParseFailed(t *testing.T) {
	incomplete := `(new a`
	_, err := Parse(strings.NewReader(incomplete))
	if err != nil {
		if _, ok := err.(*ParseError); !ok {
			t.Errorf("Parse: `%s` expecting parse error but got %s",
				incomplete, err)
		}
		return
	}
	t.Errorf("Parse `%s` is syntactically incorrect and should return error",
		incomplete)
}

// This example shows how the parser should be invoked.
func ExampleParse() {
	proc, err := Parse(strings.NewReader("(new a) (a<v> | a(x).b(y).0 | b<u>)"))
	if err != nil {
		fmt.Println(err) // Parse failed
	}
	fmt.Println(proc.String())
	// Output: restrict(a,par[ par[ send(a,[v]) | recv(a,[x]).recv(b,[y]).inact ] | send(b,[u]) ])
}
