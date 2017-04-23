package asyncpi

import (
	"fmt"
	"log"
)

type Annotater interface {
	annotate(n Name) string
}

// Uniquefier is an Annotater to test binding.
type Uniquefier struct {
	names map[Name]string
}

func (u *Uniquefier) annotate(n Name) string {
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

func AnnotateName(proc Process, a Annotater) {
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
				n.setName(a.annotate(p.Chan))
			}
			for i := range p.Vars {
				if n, ok := p.Vars[i].(nameSetter); ok {
					n.setName(a.annotate(p.Vars[i]))
				}
			}
			procs = append(procs, p.Cont)
		case *Send:
			if n, ok := p.Chan.(nameSetter); ok {
				n.setName(a.annotate(p.Chan))
			}
			for i := range p.Vals {
				if n, ok := p.Vals[i].(nameSetter); ok {
					n.setName(a.annotate(p.Vals[i]))
				}
			}
		case *Restrict:
			if n, ok := p.Name.(nameSetter); ok {
				n.setName(a.annotate(p.Name))
			}
			procs = append(procs, p.Proc)
		default:
			log.Fatal(ErrUnknownProcType{Caller: "AnnotateName", Proc: p})
		}
	}
}
