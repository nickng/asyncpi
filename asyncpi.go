// +build go1.8

package asyncpi

import (
	"bytes"
	"fmt"
	"sort"
)

// Name is channel or value.
type Name interface {
	Ident() string
	String() string
}

type names []Name

func (n names) Less(i, j int) bool { return n[i].Ident() < n[j].Ident() }

// remDup removes duplicate Names from sorted []Name.
func remDup(names []Name) []Name {
	m := make(map[string]bool)
	for _, name := range names {
		if _, seen := m[name.Ident()]; !seen {
			names[len(m)] = name
			m[name.Ident()] = true
		}
	}
	return names[:len(m)]
}

// Process is process prefixed with action.
type Process interface {
	FreeNames() []Name
	FreeVars() []Name

	// Calculi returns the calculi representation.
	Calculi() string
	String() string
}

// NilProcess is the inaction process.
type NilProcess struct{}

// NewNilProcess creates a new inaction process.
func NewNilProcess() *NilProcess {
	return new(NilProcess)
}

// FreeNames of NilProcess is defined to be empty.
func (n *NilProcess) FreeNames() []Name {
	return []Name{}
}

// FreeVars of NilProcess is defined to be empty.
func (n *NilProcess) FreeVars() []Name {
	return []Name{}
}

func (n *NilProcess) String() string {
	return "inact"
}

// Par is parallel composition of P and Q.
type Par struct {
	Procs []Process
}

// NewPar creates a new parallel composition.
func NewPar(P, Q Process) *Par { return &Par{Procs: []Process{P, Q}} }

// FreeNames of Par is the free names of composed processes.
func (p *Par) FreeNames() []Name {
	var fn []Name
	for _, proc := range p.Procs {
		fn = append(fn, proc.FreeNames()...)
	}
	sort.Slice(fn, names(fn).Less)
	return remDup(fn)
}

// FreeVars of Par is the free names of composed processes.
func (p *Par) FreeVars() []Name {
	var fv []Name
	for _, proc := range p.Procs {
		fv = append(fv, proc.FreeVars()...)
	}
	sort.Slice(fv, names(fv).Less)
	return remDup(fv)
}

func (p *Par) String() string {
	var buf bytes.Buffer
	buf.WriteString("par[ ")
	for i, proc := range p.Procs {
		if i != 0 {
			buf.WriteString(" | ")
		}
		buf.WriteString(proc.String())
	}
	buf.WriteString(" ]")
	return buf.String()
}

// Recv is input of Vars on channel Chan, with continuation Cont.
type Recv struct {
	Chan Name    // Channel to receive from.
	Vars []Name  // Variable expressions.
	Cont Process // Continuation.
}

// NewRecv creates a new Recv with given channel.
func NewRecv(u Name, P Process) *Recv {
	return &Recv{Chan: u, Cont: P}
}

// SetVars give name to what is received.
func (r *Recv) SetVars(vars []Name) {
	r.Vars = vars
}

// FreeNames of Recv is the channel and FreeNames of the continuation.
func (r *Recv) FreeNames() []Name {
	var fn []Name
	fn = append(fn, FreeNames(r.Chan)...)
	fn = append(fn, r.Cont.FreeNames()...)
	sort.Slice(fn, names(fn).Less)

	for _, v := range r.Vars {
		for i, n := range fn {
			if IsSameName(v, n) {
				fn = append(fn[:i], fn[i+1:]...)
				break // next v
			}
		}
	}
	return remDup(fn)
}

// FreeVars of Recv is the channel and FreeVars in continuation minus received variables.
func (r *Recv) FreeVars() []Name {
	var fv []Name
	fv = append(fv, r.Cont.FreeVars()...)
	sort.Slice(fv, names(fv).Less)

	ffv := fv[:0] // filtered
	for i, j := 0, 0; i < len(fv); i++ {
		for j < len(r.Vars) && r.Vars[j].Ident() < fv[i].Ident() {
			j++
		}
		if j < len(r.Vars) && r.Vars[j].Ident() != fv[i].Ident() { // overshoot
			ffv = append(ffv, fv[i])
		} else if i >= len(r.Vars) { // remaining
			ffv = append(ffv, fv[i])
		}
	}
	ffv = append(ffv, FreeVars(r.Chan)...)
	sort.Slice(ffv, names(ffv).Less)
	return remDup(ffv)
}

func (r *Recv) String() string {
	return fmt.Sprintf("recv(%s,%s).%s", r.Chan.Ident(), r.Vars, r.Cont)
}

// Repeat is a replicated Process.
type Repeat struct {
	Proc Process
}

// NewRepeat creates a new replicated process.
func NewRepeat(P Process) *Repeat {
	return &Repeat{Proc: P}
}

// FreeNames of Repeat are FreeNames in Proc.
func (r *Repeat) FreeNames() []Name {
	return r.Proc.FreeNames()
}

// FreeVars of Repeat are FreeVars in Proc.
func (r *Repeat) FreeVars() []Name {
	return r.Proc.FreeVars()
}

func (r *Repeat) String() string {
	return fmt.Sprintf("repeat(%s)", r.Proc)
}

// Restrict is scope of Process.
type Restrict struct {
	Name Name
	Proc Process
}

// NewRestricts creates consecutive restrictions from a slice of Names.
func NewRestricts(a []Name, p Process) *Restrict {
	cont := p
	for i := len(a) - 1; i >= 0; i-- {
		cont = &Restrict{Name: a[i], Proc: cont}
	}
	return cont.(*Restrict)
}

// NewRestrict creates a new restriction.
func NewRestrict(a Name, P Process) *Restrict {
	return &Restrict{Name: a, Proc: P}
}

// FreeNames of Restrict are FreeNames in Proc excluding Name.
func (r *Restrict) FreeNames() []Name {
	var fn []Name
	fn = append(fn, r.Proc.FreeNames()...)
	sort.Slice(fn, names(fn).Less)
	fn = remDup(fn)

	for i, n := range fn {
		if IsSameName(n, r.Name) {
			fn = append(fn[:i], fn[i+1:]...)
			break
		}
	}
	return fn
}

// FreeVars of Restrict are FreeVars in Proc.
func (r *Restrict) FreeVars() []Name {
	return r.Proc.FreeVars()
}

func (r *Restrict) String() string {
	return fmt.Sprintf("restrict(%s,%s)", r.Name, r.Proc)
}

// Send is output of Vals on channel Chan.
type Send struct {
	Chan Name   // Channel to send to.
	Vals []Name // Values to send.
}

// NewSend creates a new Send with given channel.
func NewSend(u Name) *Send {
	return &Send{Chan: u}
}

// SetVals determine what to send.
func (s *Send) SetVals(vals []Name) {
	s.Vals = vals
}

// FreeNames of Send is the channel and the Vals.
func (s *Send) FreeNames() []Name {
	var fn []Name
	fn = append(fn, FreeNames(s.Chan)...)
	for _, v := range s.Vals {
		fn = append(fn, FreeNames(v)...)
	}
	sort.Slice(fn, names(fn).Less)
	return remDup(fn)
}

// FreeVars of Send is the Vals.
func (s *Send) FreeVars() []Name {
	var fv []Name
	fv = append(fv, FreeVars(s.Chan)...)
	for _, v := range s.Vals {
		fv = append(fv, FreeVars(v)...)
	}
	sort.Slice(fv, names(fv).Less)
	return remDup(fv)
}

func (s *Send) String() string {
	return fmt.Sprintf("send(%s,%s)", s.Chan.Ident(), s.Vals)
}
