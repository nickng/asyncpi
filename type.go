package asyncpi

// Type system.
// A mini type system to represent types and perform type inference.

import (
	"bytes"
	"fmt"
	"log"
)

// Type is a representation of types.
type Type interface {
	Underlying() Type
	String() string
}

// unTyped is an undefined type.
type unTyped struct{}

// NewUnTyped creates a new unTyped.
func NewUnTyped() Type {
	return &unTyped{}
}

// Underlying of unTyped is itself.
func (t *unTyped) Underlying() Type {
	return t
}

func (t *unTyped) String() string {
	return "interface{}"
}

// baseType is a concrete type.
type baseType struct {
	name string
}

// NewBaseType creates a new concrete type from string type name.
func NewBaseType(t string) Type {
	return &baseType{name: t}
}

// Underlying of baseType is itself.
func (t *baseType) Underlying() Type {
	return t
}

// String of baseType returns the type name.
func (t *baseType) String() string {
	return t.name
}

// refType is a reference to the type of a given name.
// Since names don't change but types do, we use the enclosing name as a handle.
type refType struct {
	n Name
}

// NewRefType creates a new reference type from a name.
func NewRefType(n Name) Type {
	return &refType{n: n}
}

// Underlying of a refType returns the referenced type.
func (t *refType) Underlying() Type {
	return t.n.Type()
}

// String of refType returns the type name of underlying type.
func (t *refType) String() string {
	return fmt.Sprintf("%s", t.n.Type().String())
}

// compType is a composite type.
type compType struct {
	types []Type
}

// NewCompType creates a new composite type from a list of types.
func NewCompType(t ...Type) Type {
	comp := &compType{types: []Type{}}
	comp.types = append(comp.types, t...)
	return comp
}

// Underlying of a compType returns itself.
func (t *compType) Underlying() Type {
	return t
}

// String of compType is a struct of composed types.
func (t *compType) String() string {
	var buf bytes.Buffer
	buf.WriteString("struct{")
	for i, t := range t.types {
		if i != 0 {
			buf.WriteRune(';')
		}
		buf.WriteString(fmt.Sprintf("e%d %s", i, t.String()))
	}
	buf.WriteString("}")
	return buf.String()
}

func (t *compType) Elems() []Type {
	return t.types
}

// chanType is reference type wrapped with a channel.
type chanType struct {
	T Type
}

// NewChanType creates a new channel type from an existing type.
func NewChanType(t Type) Type {
	return &chanType{T: t}
}

// Underlying of a chanType is itself.
func (t *chanType) Underlying() Type {
	return t
}

// String of refType is proxy to underlying type.
func (t *chanType) String() string {
	return fmt.Sprintf("chan %s", t.T.String())
}

// BUG(nickng) Inference may fail if type of a name is recursively defined (e.g.
// a<a> → typed chan of type(a)), printing the type will cause a stack
// overflow.

// Infer performs inline type inference for channels.
//
// Infer should be called after Bind, so the types of names inferred from
// channels can be propagated to other references bound to the same name.
func Infer(p Process) {
	switch proc := p.(type) {
	case *NilProcess:
	case *Par:
		for _, proc := range proc.Procs {
			Infer(proc)
		}
	case *Recv:
		Infer(proc.Cont)
		// But that's all we know right now.
		if _, ok := proc.Chan.Type().(*unTyped); ok {
			switch arity := len(proc.Vars); arity {
			case 1:
				if t, ok := proc.Vars[0].Type().(*refType); ok { // Already a ref
					proc.Chan.SetType(NewChanType(t))
				} else {
					proc.Chan.SetType(NewChanType(NewRefType(proc.Vars[0])))
				}
			default:
				ts := []Type{}
				for i := range proc.Vars {
					if t, ok := proc.Vars[i].Type().(*refType); ok {
						ts = append(ts, t)
					} else {
						ts = append(ts, NewRefType(proc.Vars[i]))
					}
				}
				proc.Chan.SetType(NewChanType(NewCompType(ts...)))
			}
		}
	case *Send: // Send is the only place we can infer channel type.
		switch arity := len(proc.Vals); arity {
		case 1:
			if t, ok := proc.Vals[0].Type().(*refType); ok { // Already a ref
				proc.Chan.SetType(NewChanType(t))
			} else {
				proc.Chan.SetType(NewChanType(NewRefType(proc.Vals[0])))
			}
		default:
			ts := []Type{}
			for i := range proc.Vals {
				if t, ok := proc.Vals[i].Type().(*refType); ok {
					ts = append(ts, t)
				} else {
					ts = append(ts, NewRefType(proc.Vals[i]))
				}
			}
			proc.Chan.SetType(NewChanType(NewCompType(ts...)))
		}
	case *Repeat:
		Infer(proc.Proc)
	case *Restrict:
		Infer(proc.Proc)
	default:
		log.Fatal(ErrUnknownProcType{Caller: "Infer", Proc: proc})
	}
}

// Unify takes sending channel and receiving channels and try to 'unify' the
// types with best effort.
//
// One of the assumption is send and receive names are already typed as channels.
// A well typed Process should have no conflict of types during unification.
func Unify(p Process) error {
	switch proc := p.(type) {
	case *NilProcess, *Send: // No continuation.
	case *Par:
		for _, proc := range proc.Procs {
			if err := Unify(proc); err != nil {
				return err
			}
		}
	case *Recv:
		// chType is either
		// - a compType with refType fields (including struct{})
		// - a refType (non-tuple)
		chType := proc.Chan.Type().(*chanType).T
		switch arity := len(proc.Vars); arity {
		case 1:
			if _, ok := chType.(*refType); !ok {
				return &ErrTypeArity{
					Got:      len(chType.(*compType).types),
					Expected: 1,
					Msg:      fmt.Sprintf("Types from channel %s and vars have different arity", proc.Chan.Name()),
				}
			}
			if _, ok := proc.Vars[0].Type().(*unTyped); ok {
				proc.Vars[0].SetType(chType) // Chan type --> Val type.
			} else if _, ok := chType.(*refType).n.Type().(*unTyped); ok {
				chType.(*refType).n.SetType(proc.Vars[0].Type()) // Val --> Chan type
			} else if equalType(chType, proc.Vars[0].Type()) {
				// Type is both set but equal.
			} else {
				return &ErrType{
					T:   chType,
					U:   proc.Vars[0].Type(),
					Msg: fmt.Sprintf("Types inferred from channel %s are in conflict", proc.Chan.Name()),
				}
			}
		default:
			if ct, ok := chType.(*compType); !ok {
				return &ErrTypeArity{
					Got:      1,
					Expected: len(proc.Vars),
					Msg:      fmt.Sprintf("Types from channel %s and vars have different arity", proc.Chan.Name()),
				}
			} else if len(ct.types) != len(proc.Vars) {
				return &ErrTypeArity{
					Got:      len(ct.types),
					Expected: len(proc.Vars),
					Msg:      fmt.Sprintf("Types from channel %s and vars have different arity", proc.Chan.Name()),
				}
			}
			for i := range proc.Vars {
				if _, ok := proc.Vars[i].Type().(*unTyped); ok {
					proc.Vars[i].SetType(chType.(*compType).types[i].(*refType).n.Type())
				} else if _, ok := chType.(*compType).types[i].(*refType).n.Type().(*unTyped); ok {
					chType.(*compType).types[i].(*refType).n.SetType(proc.Vars[i].Type())
				} else if equalType(chType.(*compType).types[i], proc.Vars[i].Type()) {
					// Type is both set but equal.
				} else {
					return &ErrType{
						T:   chType,
						U:   proc.Vars[0].Type(),
						Msg: fmt.Sprintf("Types inferred from channel %s are in conflict", proc.Chan.Name()),
					}
				}
			}
		}
		return Unify(proc.Cont)
	case *Repeat:
		return Unify(proc.Proc)
	case *Restrict:
		return Unify(proc.Proc)
	}
	return nil
}

// ProcTypes returns the Type of the Process p.
func ProcTypes(p Process) string {
	switch proc := p.(type) {
	case *NilProcess:
		return "0"
	case *Send:
		return fmt.Sprintf("%s!%#v", proc.Chan.Name(), proc.Chan.Type())
	case *Recv:
		return fmt.Sprintf("%s?%#v; %s", proc.Chan.Name(), proc.Chan.Type(), ProcTypes(proc.Cont))
	case *Par:
		var buf bytes.Buffer
		for i, ps := range proc.Procs {
			if i != 0 {
				buf.WriteRune('|')
			}
			buf.WriteString(ProcTypes(ps))
		}
		return buf.String()
	case *Repeat:
		return "*" + ProcTypes(proc.Proc)
	case *Restrict:
		return fmt.Sprintf("(ν%s:%s) %s", proc.Name.Name(), proc.Name.Type(), ProcTypes(proc.Proc))
	default:
		log.Fatal(ErrUnknownProcType{Caller: "ProcTypes", Proc: proc})
	}
	return ""
}

// deref peels off layers of refType from a given type and returns the underlying
// type.
func deref(t Type) Type {
	if rt, ok := t.(*refType); ok {
		return deref(rt.n.Type())
	}
	return t
}

// equalType compare types.
func equalType(t, u Type) bool {
	if baseT, tok := deref(t).(*baseType); tok {
		if baseU, uok := deref(u).(*baseType); uok {
			return baseT.name == baseU.name
		}
	}
	if compT, tok := deref(t).(*compType); tok {
		if compU, uok := deref(u).(*compType); uok {
			if len(compT.types) == 0 && len(compU.types) == 0 {
				return true
			}
			compEqual := len(compT.types) == len(compU.types)
			for i := range compT.types {
				compEqual = compEqual && equalType(compT.types[i], compU.types[i])
			}
			return compEqual
		}
	}
	if chanT, tok := deref(t).(*chanType); tok {
		if chanU, uok := deref(u).(*chanType); uok {
			return equalType(chanT.T, chanU.T)
		}
	}
	return false
}
