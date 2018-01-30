package asyncpi

import (
	"fmt"
	"sort"
	"strings"
	"testing"
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
			Input: `b().a().0 | b<> | (new x)x(a,b,c).0`,
			Output: `((recv(b,[]).recv(a,[]).inact
|send(b,[]))
|restrict(x,recv(x,[a b c]).inact))`, FreeNames: newPiNames("a", "b"),
		},
		"Recv": {
			Input:     `a(b, c,d__).   0 `,
			Output:    `recv(a,[b c d__]).inact`,
			FreeNames: newPiNames("a"),
		},
		"Rep": {
			Input:     `! a().0`,
			Output:    `repeat(recv(a,[]).inact)`,
			FreeNames: []Name{newPiName("a")},
		},
		"Res": {
			Input:     `(new x)  x().0 `,
			Output:    `restrict(x,recv(x,[]).inact)`,
			FreeNames: []Name{},
		},
		"Send": {
			Input:     `a<b, e_, b> `,
			Output:    `send(a,[b e_ b])`,
			FreeNames: []Name{newPiName("a"), newPiName("b"), newPiName("e_")},
		},
	}
}

// Tests fn(a) is a
func TestFreeName(t *testing.T) {
	name := newPiName("a")
	freeNames := name.FreeNames()
	if len(freeNames) == 1 && freeNames[0].Name() != name.Name() {
		t.Errorf("FreeName: fn(a) does not match a: `%s` vs `%s`", freeNames, name)
	}
}

// Tests fn(a) U fn(b) is a U b
func TestFreeNames(t *testing.T) {
	piNames := []Name{newPiName("a"), newPiName("c"), newPiName("b")}
	freeNames := []Name{}
	for _, name := range piNames {
		freeNames = append(freeNames, name.FreeNames()...)
	}
	sort.Slice(freeNames, names(freeNames).Less)
	sort.Slice(piNames, names(piNames).Less)
	if len(piNames) != len(freeNames) {
		t.Errorf("FreeNames: fn(a...) and a... have different sizes")
		t.Fail()
	}
	for i := range piNames {
		if piNames[i].Name() != freeNames[i].Name() {
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
		if fn.Name() != test.FreeNames[i].Name() {
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
	// Output: restrict(a,((send(a,[v])
	// |recv(a,[x]).recv(b,[y]).inact)
	// |send(b,[u])))
}
