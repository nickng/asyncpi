package asyncpi

import (
	"bytes"
	"fmt"
	"log"
)

// name is a concrete Name.
type piName struct {
	name string
	s    sorts
}

// newPiName creates a new concrete name from a string.
func newPiName(n string) *piName {
	return &piName{name: n}
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

func (n *piName) String() string {
	var buf bytes.Buffer
	if n.s == varSort {
		buf.WriteString("_")
	}
	buf.WriteString(n.name)
	return buf.String()
}

// hintedName represents a Name extended with type hint.
type hintedName struct {
	name Name
	hint string
}

// newHintedName returns a new Name for the given n and t.
func newHintedName(name Name, hint string) *hintedName {
	return &hintedName{name, hint}
}

func (n *hintedName) FreeNames() []Name {
	return n.name.FreeNames()
}

func (n *hintedName) FreeVars() []Name {
	return n.name.FreeVars()
}

func (n *hintedName) Name() string {
	return n.name.Name()
}

// TypeHint returns the type hint attached to given Name.
func (n *hintedName) TypeHint() string {
	return n.hint
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
	s := fmt.Sprintf("%s_%d", n.Name(), len(u.names))
	u.names[n] = s
	return s
}

type nameSetter interface {
	setName(string)
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
			if n, ok := p.Chan.(nameSetter); ok {
				n.setName(a.visit(p.Chan))
			} else {
				return ImmutableNameError{Name: p.Chan}
			}
			for i := range p.Vars {
				if n, ok := p.Vars[i].(nameSetter); ok {
					n.setName(a.visit(p.Vars[i]))
				} else {
					return ImmutableNameError{Name: p.Vars[i]}
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
				} else {
					return ImmutableNameError{Name: p.Vals[i]}
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
	return nil
}
