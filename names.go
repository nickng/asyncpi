package asyncpi

import (
	"fmt"
	"go.nickng.io/asyncpi/internal/name"
	"log"
)

// newNames is a convenient utility function
// for creating a []Name from given strings.
func newNames(names ...string) []Name {
	pn := make([]Name, len(names))
	for i, n := range names {
		pn[i] = name.New(n)
	}
	return pn
}

type TypeHinter interface {
	TypeHint() string
}

type NameVisitor interface {
	visit(n Name) string
}

// Uniquefier is NameVisitor to test binding.
type Uniquefier struct {
	names map[Name]string
}

func (u *Uniquefier) visit(n Name) string {
	if u.names == nil {
		u.names = make(map[Name]string)
	}
	if un, ok := u.names[n]; ok {
		return un
	}
	s := fmt.Sprintf("%s_%d", n.Ident(), len(u.names))
	u.names[n] = s
	return s
}

type NameSetter interface {
	SetName(string)
}

func UpdateName(proc Process, a NameVisitor) error {
	procs := []Process{proc}
	for len(procs) > 0 {
		p := procs[0]
		procs = procs[1:]
		switch p := p.(type) {
		case *NilProcess:
		case *Repeat:
			procs = append(procs, p.Proc)
		case *Par:
			procs = append(procs, p.Procs...)
		case *Recv:
			if n, ok := p.Chan.(NameSetter); ok {
				n.SetName(a.visit(p.Chan))
			} else {
				return ImmutableNameError{Name: p.Chan}
			}
			for i := range p.Vars {
				if n, ok := p.Vars[i].(NameSetter); ok {
					n.SetName(a.visit(p.Vars[i]))
				} else {
					return ImmutableNameError{Name: p.Vars[i]}
				}
			}
			procs = append(procs, p.Cont)
		case *Send:
			if n, ok := p.Chan.(NameSetter); ok {
				n.SetName(a.visit(p.Chan))
			}
			for i := range p.Vals {
				if n, ok := p.Vals[i].(NameSetter); ok {
					n.SetName(a.visit(p.Vals[i]))
				} else {
					return ImmutableNameError{Name: p.Vals[i]}
				}
			}
		case *Restrict:
			if n, ok := p.Name.(NameSetter); ok {
				n.SetName(a.visit(p.Name))
			}
			procs = append(procs, p.Proc)
		default:
			log.Fatal(UnknownProcessTypeError{Caller: "UpdateName", Proc: p})
		}
	}
	return nil
}

// freeNameser is an interface which Name should
// provide to have custom FreeNames implementation.
type freeNameser interface {
	FreeNames() []Name
}

// FreeNames returns the free names in a give Name n.
func FreeNames(n Name) []Name {
	if fn, ok := n.(freeNameser); ok {
		return fn.FreeNames()
	}
	return []Name{n}
}

// freeVarser is an interface which Name should
// provide to have custom FreeVars implementation.
type freeVarser interface {
	FreeVars() []Name
}

// FreeVars returns the free variables in a give Name n.
func FreeVars(n Name) []Name {
	if fv, ok := n.(freeVarser); ok {
		return fv.FreeVars()
	}
	return nil
}

// IsSameName is a simple comparison operator for Name.
// A Name x is equal with another Name y when x and y has the same name.
// This comparison ignores the underlying representation (sort, type, etc.).
func IsSameName(x, y Name) bool {
	return x.Ident() == y.Ident()
}

func IsFreeName(x Name) bool {
	return len(FreeNames(x)) == 1 && FreeNames(x)[0].Ident() == x.Ident()
}
