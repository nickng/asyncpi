package asyncpi

import (
	"bytes"
	"fmt"
	"log"
)

// name is a concrete Name.
type piName struct {
	name string
	t    Type
	s    sorts
}

// newPiName creates a new concrete name from a string.
func newPiName(n string) Name {
	return &piName{name: n, t: NewUnTyped()}
}

// newTypedPiName creates a new concrete name with a type hint.
func newTypedPiName(n, t string) Name {
	return &piName{name: n, t: NewBaseType(t)}
}

// setSort sets the name sort.
func (n *piName) setSort(s sorts) {
	n.s = s
}

// setName sets the internal name.
func (n *piName) setName(name string) {
	n.name = name
}

// FreeNames of name is itself (if sort is name).
func (n *piName) FreeNames() []Name {
	if n.s == nameSort {
		return []Name{n}
	}
	return []Name{}
}

// FreeVars of name is itself (if sort is var).
func (n *piName) FreeVars() []Name {
	if n.s == varSort {
		return []Name{n}
	}
	return []Name{}
}

// Name is the string identifier of a name.
func (n *piName) Name() string {
	return n.name
}

// Type is the defined typed of the name.
func (n *piName) Type() Type {
	return n.t
}

// SetType sets the type of the name.
func (n *piName) SetType(t Type) {
	n.t = t
}

func (n *piName) String() string {
	var buf bytes.Buffer
	if n.s == varSort {
		buf.WriteString("_")
	}
	if _, ok := n.t.(*unTyped); ok {
		buf.WriteString(n.name)
		return buf.String()
	}
	buf.WriteString(fmt.Sprintf("%s:%s", n.name, n.t.String()))
	return buf.String()
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
	s := fmt.Sprintf("%s_%d", n.Name(), len(u.names))
	u.names[n] = s
	return s
}

type nameSetter interface {
	setName(string)
}

func UpdateName(proc Process, a NameVisitor) {
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
			if n, ok := p.Chan.(nameSetter); ok {
				n.setName(a.visit(p.Chan))
			}
			for i := range p.Vars {
				if n, ok := p.Vars[i].(nameSetter); ok {
					n.setName(a.visit(p.Vars[i]))
				}
			}
			procs = append(procs, p.Cont)
		case *Send:
			if n, ok := p.Chan.(nameSetter); ok {
				n.setName(a.visit(p.Chan))
			}
			for i := range p.Vals {
				if n, ok := p.Vals[i].(nameSetter); ok {
					n.setName(a.visit(p.Vals[i]))
				}
			}
		case *Restrict:
			if n, ok := p.Name.(nameSetter); ok {
				n.setName(a.visit(p.Name))
			}
			procs = append(procs, p.Proc)
		default:
			log.Fatal(UnknownProcessTypeError{Caller: "UpdateName", Proc: p})
		}
	}
}
