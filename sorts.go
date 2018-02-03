package asyncpi

import (
	"log"
	"strings"
	"unicode/utf8"
)

type sorts int

const (
	nameSort sorts = iota // default sort is name.
	varSort
)

// Sorts.
// Names are split to name and var sort.

// SortedName represents a Name with sort.
type SortedName struct {
	Name
	s sorts
}

// NameWithSort returns a new SortedName based on the given Name n.
func NameWithSort(n Name, s sorts) *SortedName {
	return &SortedName{n, s}
}

func (n *SortedName) FreeNames() []Name {
	if n.s == nameSort {
		return []Name{n}
	}
	return nil
}

func (n *SortedName) FreeVars() []Name {
	if n.s == varSort {
		return []Name{n}
	}
	return nil
}

// Sort returns the sort of the Name n.
func (n *SortedName) Sort() sorts {
	return n.s
}

// setSort sets the sort of the given sorted Name n.
func (n *SortedName) SetSort(s sorts) {
	n.s = s
}

type SortSetter interface {
	SetSort(s sorts)
}

// IdentifySort puts names in a Process into their respective sort {name,var}.
func IdentifySorts(proc Process) {
	nameVar := make(map[Name]bool)
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
			for _, v := range p.Vars {
				nameVar[v] = true
				if s, ok := v.(SortSetter); ok {
					s.SetSort(varSort)
				}
			}
			procs = append(procs, p.Cont)
		case *Send:
			for _, v := range p.Vals {
				if _, ok := nameVar[v]; !ok {
					nameVar[v] = true
					if s, ok := v.(SortSetter); ok {
						s.SetSort(varSort)
					}
				}
			}
		case *Restrict:
			nameVar[p.Name] = false // new name = not var
			procs = append(procs, p.Proc)
		default:
			log.Fatal(UnknownProcessTypeError{Caller: "IdentifySorts", Proc: p})
		}
	}
}

func ResetSorts(proc Process) {
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
			if s, ok := p.Chan.(SortSetter); ok {
				s.SetSort(nameSort)
			}
			for _, v := range p.Vars {
				if s, ok := v.(SortSetter); ok {
					s.SetSort(nameSort)
				}
			}
			procs = append(procs, p.Cont)
		case *Send:
			if s, ok := p.Chan.(SortSetter); ok {
				s.SetSort(nameSort)
			}
			for _, v := range p.Vals {
				if s, ok := v.(SortSetter); ok {
					s.SetSort(nameSort)
				}
			}
		case *Restrict:
			procs = append(procs, p.Proc)
		default:
			log.Fatal(UnknownProcessTypeError{Caller: "ResetSorts", Proc: p})
		}
	}
}

// NameVarSorter is a name visitor which puts names in sorts.
// A Name is a name/var depending on its prefix:
//   names={a,b,c,...} vars={...,x,y,z}
//
type NameVarSorter struct{}

func (s *NameVarSorter) visit(n Name) string {
	r, _ := utf8.DecodeRuneInString(n.Ident())
	if strings.ContainsRune("nopqrstuvwxyz", r) {
		if s, ok := n.(SortSetter); ok {
			s.SetSort(varSort)
		}
	}
	return n.Ident()
}
