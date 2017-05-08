package asyncpi

import (
	"log"
	"strings"
	"unicode/utf8"
)

// Sorts.
// Names are split to name and var sort.

type sortSetter interface {
	setSort(s sorts)
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
				if s, ok := v.(sortSetter); ok {
					s.setSort(varSort)
				}
			}
			procs = append(procs, p.Cont)
		case *Send:
			for _, v := range p.Vals {
				if _, ok := nameVar[v]; !ok {
					nameVar[v] = true
					if s, ok := v.(sortSetter); ok {
						s.setSort(varSort)
					}
				}
			}
		case *Restrict:
			nameVar[p.Name] = false // new name = not var
			procs = append(procs, p.Proc)
		default:
			log.Fatal(ErrUnknownProcType{Caller: "IdentifySorts", Proc: p})
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
			if s, ok := p.Chan.(sortSetter); ok {
				s.setSort(nameSort)
			}
			for _, v := range p.Vars {
				if s, ok := v.(sortSetter); ok {
					s.setSort(nameSort)
				}
			}
			procs = append(procs, p.Cont)
		case *Send:
			if s, ok := p.Chan.(sortSetter); ok {
				s.setSort(nameSort)
			}
			for _, v := range p.Vals {
				if s, ok := v.(sortSetter); ok {
					s.setSort(nameSort)
				}
			}
		case *Restrict:
			procs = append(procs, p.Proc)
		default:
			log.Fatal(ErrUnknownProcType{Caller: "ResetSorts", Proc: p})
		}
	}
}

// NameVarSorter is a name visitor which puts names in sorts.
// A Name is a name/var depending on its prefix:
//   names={a,b,c,...} vars={...,x,y,z}
//
type NameVarSorter struct{}

func (s *NameVarSorter) annotate(n Name) string {
	r, _ := utf8.DecodeRuneInString(n.Name())
	if strings.ContainsRune("nopqrstuvwxyz", r) {
		if s, ok := n.(sortSetter); ok {
			s.setSort(varSort)
		}
	}
	return n.Name()
}
