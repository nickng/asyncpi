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
		"NilProcess": TestCase{
			Input:     `    0 `,
			Output:    `Inaction`,
			FreeNames: []Name{},
		},
		"Par": TestCase{
			Input: `b().a().0 | b<> | (new x)x(a,b,c).0`,
			Output: `Recv(b, [])
Recv(a, [])
Inaction
--- parallel ---
Send(b, [])
--- parallel ---
scope x {
Recv(x, [a b c])
Inaction
}`, FreeNames: []Name{newPiName("a"), newPiName("b")},
		},
		"Recv": TestCase{
			Input: `a(b, c,d__).   0 `,
			Output: `Recv(a, [b c d__])
Inaction`,
			FreeNames: []Name{newPiName("a")},
		},
		"Rep": TestCase{
			Input: `! a().0`,
			Output: `repeat {
Recv(a, [])
Inaction
}`,
			FreeNames: []Name{newPiName("a")},
		},
		"Res": TestCase{
			Input: `(new x)  x().0 `,
			Output: `scope x {
Recv(x, [])
Inaction
}`,
			FreeNames: []Name{},
		},
		"Send": TestCase{
			Input:     `a<b, e_, b> `,
			Output:    `Send(a, [b e_ b])`,
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
	names := []Name{newPiName("a"), newPiName("c"), newPiName("b")}
	freeNames := []Name{}
	for _, name := range names {
		freeNames = append(freeNames, name.FreeNames()...)
	}
	sort.Sort(byName(freeNames))
	sort.Sort(byName(names))
	if len(names) != len(freeNames) {
		t.Errorf("FreeNames: fn(a...) and a... have different sizes")
		t.Fail()
	}
	for i := range names {
		if names[i].Name() != freeNames[i].Name() {
			t.Errorf("FreeNames: fn(a...) does not match a...: `%s` vs `%s`", freeNames, names)
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
		if _, ok := err.(*ErrParse); !ok {
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
	// Output: scope a {
	//Send(a, [v])
	//--- parallel ---
	//Recv(a, [x])
	//Recv(b, [y])
	//Inaction
	//--- parallel ---
	//Send(b, [u])
	//}
}

// This example shows how to generate code from a Process.
/*
func ExampleCodegen() {
	proc, err := Parse(strings.NewReader("(new a)(new b)(a<b> | a(x).x<> | b())"))
	if err != nil {
		fmt.Println(err) // Parse failed
	}
	fmt.Println(Codegen(proc))
	// Output: ""
}
*/
