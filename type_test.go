package asyncpi

import (
	"strings"
	"testing"
)

// Tests type inference only.
func TestBasicInferOnly(t *testing.T) {
	input := "(new a)(new b)(new c:T)(a<b,c>|a(y,z).b<z>)"
	atype := "chan struct{e0 chan interface{};e1 T}"
	btype := "chan interface{}" // chan type(c) if unified.
	ctype := "T"                // From type hint.
	ytype := "interface{}"      // type(b) if unified.
	ztype := "interface{}"      // type(c) if unified.
	proc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	proc = Bind(proc)
	Infer(proc)

	resa, ok := proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new a)P but got `%s`", proc)
	}
	if resa.Name.Type().String() != atype {
		t.Errorf("Infer: expected a typed `%s` but got `%s`",
			atype, resa.Name.Type().String())
	}
	resb, ok := resa.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new b)P but got `%s`", resa.Proc)
	}
	if resb.Name.Type().String() != btype {
		t.Errorf("Infer: expected b typed `%s` but got `%s`",
			btype, resb.Name.Type().String())
	}
	resc, ok := resb.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new c)P but got `%s`", resb.Proc)
	}
	if resc.Name.Type().String() != ctype {
		t.Errorf("Infer: expected c typed `%s` but got `%s`",
			ctype, resc.Name.Type().String())
	}
	par, ok := resc.Proc.(*Par)
	if !ok {
		t.Errorf("Parse: expected P|Q but got `%s`", resc.Proc)
	}
	recva, ok := par.Procs[1].(*Recv)
	if !ok {
		t.Errorf("Parse: expected a() but got `%s`", par.Procs[1])
	}
	if len(recva.Vars) != 2 {
		t.Errorf("Parse: expected a(y,z) but got a %d-tuple", len(recva.Vars))
	}
	if recva.Vars[0].Type().String() != ytype {
		t.Errorf("Infer: expected y typed `%s` but got `%s`",
			ytype, recva.Vars[0].Type().String())
	}
	sendb, ok := recva.Cont.(*Send)
	if !ok {
		t.Errorf("Parse: expected b<z> but got `%s`", recva.Cont)
	}
	if len(sendb.Vals) != 1 {
		t.Errorf("Parse: expected b<z> but got a %d-tuple", len(sendb.Vals))
	}
	if sendb.Vals[0].Type().String() != ztype {
		t.Errorf("Infer: expected z typed `%s` but got `%s`",
			ztype, sendb.Vals[0].Type().String())
	}
}

// Tests type inference and unification.
// Tis propages down (channel type to value types).
func TestBasicInferUnify(t *testing.T) {
	input := "(new a)(new b)(new c:T)(a<b,c>|a(y,z).b<z>)"
	atype := "chan struct{e0 chan T;e1 T}"
	btype := "chan T" // chan type(c).
	ctype := "T"      // From type hint.
	ytype := "chan T" // type(b).
	ztype := "T"      // type(c).
	proc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	proc = Bind(proc)
	Infer(proc)
	Unify(proc)

	resa, ok := proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new a)P but got `%s`", resa)
	}
	if resa.Name.Type().String() != atype {
		t.Errorf("Infer: expected a typed `%s` but got `%s`",
			atype, resa.Name.Type().String())
	}
	resb, ok := resa.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new b)P but got `%s`", resb)
	}
	if resb.Name.Type().String() != btype {
		t.Errorf("Infer: expected b typed `%s` but got `%s`",
			btype, resb.Name.Type().String())
	}
	resc, ok := resb.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new c)P but got `%s`", resc)
	}
	if resc.Name.Type().String() != ctype {
		t.Errorf("Infer: expected c typed `%s` but got `%s`",
			ctype, resc.Name.Type().String())
	}
	par, ok := resc.Proc.(*Par)
	if !ok {
		t.Errorf("Parse: expected P|Q but got `%s`", par)
	}
	recva, ok := par.Procs[1].(*Recv)
	if !ok {
		t.Errorf("Infer: expected a() but got `%s`", recva)
	}
	if len(recva.Vars) != 2 {
		t.Errorf("Parse: expected a(y,z) but got a %d-tuple", len(recva.Vars))
	}
	if recva.Vars[0].Type().String() != ytype {
		t.Errorf("Infer: expected y typed `%s` but got `%s`",
			ytype, recva.Vars[0].Type().String())
	}
	sendb, ok := recva.Cont.(*Send)
	if !ok {
		t.Errorf("Infer: expected b<z> but got `%s`", sendb)
	}
	if len(sendb.Vals) != 1 {
		t.Errorf("Parse: expected b<z> but got a %d-tuple", len(sendb.Vals))
	}
	if sendb.Vals[0].Type().String() != ztype {
		t.Errorf("Infer: expected z typed `%s` but got `%s`",
			ztype, sendb.Vals[0].Type().String())
	}
}

// Tests type inference with wrong type hint.
//
// a sends b and c so a is a 2-elem struct chan.
// b is T by type hint, so a is struct{e0 T;e1 interface{}}
func TestWrongHintInferOnly(t *testing.T) {
	simple := "(new a:TA)(new b:T)(new c)(a<b,c>|a(x,y).0)"
	atype := "chan struct{e0 T;e1 interface{}}"
	btype := "T" // From type hint.
	ctype := "interface{}"
	xtype := "interface{}" // type(b) if unified.
	ytype := "interface{}" // type(c) if unified.
	proc, err := Parse(strings.NewReader(simple))
	if err != nil {
		t.Fatal(err)
	}
	proc = Bind(proc)
	Infer(proc)

	resa, ok := proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new a)P but got `%s`", proc)
	}
	if resa.Name.Type().String() != atype {
		t.Errorf("Infer: expected a typed `%s` but got `%s`",
			atype, resa.Name.Type().String())
	}
	resb, ok := resa.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new b)P but got `%s`", resa.Proc)
	}
	if resb.Name.Type().String() != btype {
		t.Errorf("Infer: expected b typed `%s` but got `%s`",
			btype, resb.Name.Type().String())
	}
	resc, ok := resb.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse expected (new c)P but got `%s`", resb.Proc)
	}
	if resc.Name.Type().String() != ctype {
		t.Errorf("Infer: expected c typed `%s` but got `%s`",
			ctype, resc.Name.Type().String())
	}
	par, ok := resc.Proc.(*Par)
	if !ok {
		t.Errorf("Parse: expected P|Q but got `%s`", resc.Proc)
	}
	recva, ok := par.Procs[1].(*Recv)
	if !ok {
		t.Errorf("Parse: expected a(x,y) but got `%s`", par.Procs[1])
	}
	if len(recva.Vars) != 2 {
		t.Errorf("Parse: expected a(x,y) but got a %d-tuple", len(recva.Vars))
	}
	if recva.Vars[0].Type().String() != xtype {
		t.Errorf("Infer: expected y typed `%s` but got `%s`",
			xtype, recva.Vars[0].Type().String())
	}
	if recva.Vars[1].Type().String() != ytype {
		t.Errorf("Infer: expected y typed `%s` but got `%s`",
			ytype, recva.Vars[1].Type().String())
	}
}

// Tests type inference and unification with wrong type hint.
// This is a one-way propagation (channel to value).
//
// a sends b and c so a is a 2-elem struct chan.
// b is T by type hint, so a is struct{e0 T;e1 interface{}}
// x and y are determined by unification
func TestWrongHintInferUnify(t *testing.T) {
	simple := "(new a:TA)(new b:T)(new c)(a<b,c>|a(x,y).0)"
	atype := "chan struct{e0 T;e1 interface{}}"
	btype := "T" // From type hint.
	ctype := "interface{}"
	xtype := "T"           // type(b).
	ytype := "interface{}" // type(c).
	proc, err := Parse(strings.NewReader(simple))
	if err != nil {
		t.Fatal(err)
	}
	proc = Bind(proc)
	Infer(proc)
	Unify(proc)

	resa, ok := proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new a)P but got `%s`", proc)
	}
	if resa.Name.Type().String() != atype {
		t.Errorf("Infer: expected a typed `%s` but got `%s`",
			atype, resa.Name.Type().String())
	}
	resb, ok := resa.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new b)P but got `%s`", resa.Proc)
	}
	if resb.Name.Type().String() != btype {
		t.Errorf("Infer: expected b typed `%s` but got `%s`",
			btype, resb.Name.Type().String())
	}
	resc, ok := resb.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse expected (new c)P but got `%s`", resb.Proc)
	}
	if resc.Name.Type().String() != ctype {
		t.Errorf("Infer: expected c typed `%s` but got `%s`",
			ctype, resc.Name.Type().String())
	}
	par, ok := resc.Proc.(*Par)
	if !ok {
		t.Errorf("Parse: expected P|Q but got `%s`", resc.Proc)
	}
	recva, ok := par.Procs[1].(*Recv)
	if !ok {
		t.Errorf("Parse: expected a(x,y) but got `%s`", par.Procs[1])
	}
	if len(recva.Vars) != 2 {
		t.Errorf("Parse: expected a(x,y) but got a %d-tuple", len(recva.Vars))
	}
	if recva.Vars[0].Type().String() != xtype {
		t.Errorf("Infer: expected y typed `%s` but got `%s`",
			xtype, recva.Vars[0].Type().String())
	}
	if recva.Vars[1].Type().String() != ytype {
		t.Errorf("Infer: expected y typed `%s` but got `%s`",
			ytype, recva.Vars[1].Type().String())
	}
}

// Tests type inference on higher order names.
func TestHigherOrderInferOnly(t *testing.T) {
	input := "(new a)(new b)(a<b>|a(x).x().0)"
	atype := "chan interface{}" // chan type(x) if unified.
	btype := "interface{}"      // chan type(x) if unified.
	xtype := "chan struct{}"
	proc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	proc = Bind(proc)
	Infer(proc)
	resa, ok := proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new a)P but got `%s`", proc)
	}
	if resa.Name.Type().String() != atype {
		t.Errorf("Infer: expected a typed `%s` but got `%s`",
			atype, resa.Name.Type().String())
	}
	resb, ok := resa.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new b)P but got `%s`", resa.Proc)
	}
	if resb.Name.Type().String() != btype {
		t.Errorf("Infer: expected b typed `%s` but got `%s`",
			btype, resb.Name.Type().String())
	}
	par, ok := resb.Proc.(*Par)
	if !ok {
		t.Errorf("Parse: expected P|Q but got `%s`", resb.Proc)
	}
	recva, ok := par.Procs[1].(*Recv)
	if !ok {
		t.Errorf("Parse: expected a(x) but got `%s`", par.Procs[1])
	}
	if len(recva.Vars) != 1 {
		t.Errorf("Parse: expected a(x) but got a %d-tuple", len(recva.Vars))
	}
	if recva.Vars[0].Type().String() != xtype {
		t.Errorf("Infer: expected x typed `%s` but got `%s`",
			xtype, recva.Vars[0].Type().String())
	}
}

// Tests type inference and unification on higher order names.
// This propagates up (value type of b to channel).
func TestHigherOrderInferUnify(t *testing.T) {
	input := "(new a)(new b)(a<b>|a(x).x().0)"
	atype := "chan chan struct{}" // chan type(x).
	btype := "chan struct{}"      // chan type(x).
	xtype := "chan struct{}"
	proc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	proc = Bind(proc)
	Infer(proc)
	Unify(proc)
	resa, ok := proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new a)P but got `%s`", proc)
	}
	if resa.Name.Type().String() != atype {
		t.Errorf("Infer: expected a typed `%s` but got `%s`",
			atype, resa.Name.Type().String())
	}
	resb, ok := resa.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: expected (new b)P but got `%s`", resa.Proc)
	}
	if resb.Name.Type().String() != btype {
		t.Errorf("Infer: expected b typed `%s` but got `%s`",
			btype, resb.Name.Type().String())
	}
	par, ok := resb.Proc.(*Par)
	if !ok {
		t.Errorf("Parse: expected P|Q but got `%s`", resb.Proc)
	}
	recva, ok := par.Procs[1].(*Recv)
	if !ok {
		t.Errorf("Parse: expected a(x) but got `%s`", par.Procs[1])
	}
	if len(recva.Vars) != 1 {
		t.Errorf("Parse: expected a(x) but got a %d-tuple", len(recva.Vars))
	}
	if recva.Vars[0].Type().String() != xtype {
		t.Errorf("Infer: expected x typed `%s` but got `%s`",
			xtype, recva.Vars[0].Type().String())
	}
}

// Tests inference and unification of nested type.
// a sends b and c so a is 2-elem struct.
// c is T by type hint, so a is struct{e0 interface{};e1 T}
// a receives into y,z so y is interface{}, z is T
// b sends c and z so b is 2-elem struct, combined with above, struct{e0 T;e1 T}
// a is therefore struct{e0 struct{e0 T; e1 T};e1 T}
func TestInferUnifyNested(t *testing.T) {
	nested := "(new a)(new b)(new c:T)(a<b,c> | a(y,z).b<c,z>)"
	atype := "chan struct{e0 chan struct{e0 T;e1 T};e1 T}"
	btype := "chan struct{e0 T;e1 T}"
	proc, err := Parse(strings.NewReader(nested))
	if err != nil {
		t.Fatal(err)
	}
	resa, ok := proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: `%s` does not begin with restriction", nested)
	}
	if _, ok := resa.Name.Type().(*unTyped); !ok {
		t.Errorf("Infer: Type of `a` is not %s\n got: %s",
			atype, resa.Name.Type())
	}
	proc = Bind(proc)
	Infer(proc)
	Unify(proc)
	inferredResa, ok := proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: `%s` does not begin with restriction", nested)
	}
	if inferredResa.Name.Type().String() != atype {
		t.Errorf("Infer: Type of `a` is not %s\ngot: %s",
			atype, inferredResa.Name.Type())
	}
	inferredResb, ok := inferredResa.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: `%s` does not begin with restriction",
			inferredResa.Calculi())
	}
	if inferredResb.Name.Type().String() != btype {
		t.Errorf("Infer: Type of `b` is not %s\ngot: %s",
			btype, inferredResb.Name.Type())
	}
}

// Tests inference and unification of name-passing type.
// a sends b and c so a is 2-elem struct.
// c is T by type hint, so a is struct{e0 interface{};e1 T}
// a receives into y,z so y is interface{}, z is T
// y sends c and z so y is 2-elem struct, combined with above, struct{e0 T;e1 T}
// unification makes y <=> b
// a is therefore struct{e0 struct{e0 T; e1 T};e1 T}
/*
func TestInferUnifyNamePassing(t *testing.T) {
	namePassing := "(new a)(new b)(new c:T)(a<b,c> | a(y,z).y<c,z>)"
	atype := "chan struct{e0 chan struct{e0 T;e1 T};e1 T}"
	btype := "chan struct{e0 T;e1 T}"
	proc, err := Parse(strings.NewReader(namePassing))
	if err != nil {
		t.Fatal(err)
	}
	resa, ok := proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: `%s` does not begin with restriction", namePassing)
	}
	if _, ok := resa.Name.Type().(*unTyped); !ok {
		t.Errorf("Infer: Type of `a` is not %s\n got: %s",
			atype, resa.Name.Type())
	}
	proc = Bind(proc)
	Infer(proc)
	Unify(proc)
	inferredResa, ok := proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: `%s` does not begin with restriction", namePassing)
	}
	if inferredResa.Name.Type().String() != atype {
		t.Errorf("Infer: Type of `a` is not %s\ngot: %s",
			atype, inferredResa.Name.Type())
	}
	inferredResb, ok := inferredResa.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: `%s` does not begin with restriction",
			inferredResa.Calculi())
	}
	if inferredResb.Name.Type().String() != btype {
		t.Errorf("Infer: Type of `b` is not %s\ngot: %s",
			btype, inferredResb.Name.Type())
	}
}
*/

// Tests inference of nested type.
// a sends b and c so a is 2-elem struct.
// c is T by type hint, so a is struct{e0 interface{};e1 T}
// a receives into y,z so y is interface{}, z is T
// b sends c and z so b is 2-elem struct, combined with above, struct{e0 T;e1 T}
// a is therefore struct{e0 struct{e0 T; e1 T};e1 T}
func TestInferNested(t *testing.T) {
	nested := "(new a)(new b)(new c:T)(a<b,c> | a(y,z).y<z>)"
	atype := "chan struct{e0 chan T;e1 T}"
	btype := "chan T"
	proc, err := Parse(strings.NewReader(nested))
	if err != nil {
		t.Fatal(err)
	}
	resa, ok := proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: `%s` does not begin with restriction", nested)
	}
	if _, ok := resa.Name.Type().(*unTyped); !ok {
		t.Errorf("Infer: Type of `a` is not %s\n got: %s",
			atype, resa.Name.Type())
	}
	proc = Bind(proc)
	Infer(proc)
	Unify(proc)
	inferredResa, ok := proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: `%s` does not begin with restriction", nested)
	}
	if inferredResa.Name.Type().String() != atype {
		t.Errorf("Infer: Type of `a` is not %s\ngot: %s",
			atype, inferredResa.Name.Type())
	}
	inferredResb, ok := inferredResa.Proc.(*Restrict)
	if !ok {
		t.Errorf("Parse: `%s` does not begin with restriction",
			inferredResa.Calculi())
	}
	if inferredResb.Name.Type().String() != btype {
		t.Errorf("Infer: Type of `b` is not %s\ngot: %s",
			btype, inferredResb.Name.Type())
	}
}

// Test mismatched comptype (different number of args).
// a,b are concrete, and a is bound. a cannot have both 2 args and 0 args.
// both a are compType because of the binding, but the args mismatches.
func TestMismatchCompType(t *testing.T) {
	incompat := `(new a,b)(a(b,c).0 | a<>)`
	proc, err := Parse(strings.NewReader(incompat))
	if err != nil {
		t.Fatal(err)
	}
	bproc := Bind(proc)
	Infer(bproc)
	err = Unify(bproc)
	if _, ok := err.(*TypeArityError); !ok {
		t.Fatalf("Unify: Expecting type error (mismatched args in a) but got", err)
	}
}

// Test receive with multiple compatible senders.
func TestMultipleSender(t *testing.T) {
	ms := `(new b)a(x).(x<b>|x<z>)`
	proc, err := Parse(strings.NewReader(ms))
	if err != nil {
		t.Fatal(err)
	}
	bproc := Bind(proc)
	Infer(bproc)
	if err := Unify(bproc); err != nil {
		t.Fatal(err)
	}
}
