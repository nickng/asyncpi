// +build go1.8

package asyncpi

import (
	"bytes"
	"fmt"
	"log"
	"sort"

	"text/template"
)

// Name is channel or value.
type Name interface {
	FreeNames() []Name
	FreeVars() []Name
	Name() string
}

type names []Name

func (n names) Less(i, j int) bool { return n[i].Name() < n[j].Name() }

// remDup removes duplicate Names from sorted []Name.
func remDup(names []Name) []Name {
	m := make(map[string]bool)
	for _, name := range names {
		if _, seen := m[name.Name()]; !seen {
			names[len(m)] = name
			m[name.Name()] = true
		}
	}
	return names[:len(m)]
}

// Process is process prefixed with action.
type Process interface {
	FreeNames() []Name
	FreeVars() []Name

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

// Calculi returns the calculi representation.
func (n *NilProcess) Calculi() string {
	return "0"
}

func (n *NilProcess) String() string {
	return "Inaction\n"
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

// Calculi returns the calculi representation.
func (p *Par) Calculi() string {
	buf := new(bytes.Buffer)
	t := template.Must(template.New("par").Parse(`(
{{- range $i, $p := .Procs -}}
{{- if $i }} | {{ end -}}{{- $p.Calculi -}}
{{- end -}})`))
	err := t.Execute(buf, p)
	if err != nil {
		log.Println("Executing template:", err)
		return ""
	}
	return buf.String()
}

func (p *Par) String() string {
	var buf bytes.Buffer
	for i, proc := range p.Procs {
		if i != 0 {
			buf.WriteString("--- parallel ---\n")
		}
		buf.WriteString(proc.String())
	}
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
	fn = append(fn, r.Chan.FreeNames()...)
	fn = append(fn, r.Cont.FreeNames()...)
	sort.Slice(fn, names(fn).Less)
	return remDup(fn)
}

// FreeVars of Recv is the channel and FreeVars in continuation minus received variables.
func (r *Recv) FreeVars() []Name {
	var fv []Name
	fv = append(fv, r.Cont.FreeVars()...)
	sort.Slice(fv, names(fv).Less)

	ffv := fv[:0] // filtered
	for i, j := 0, 0; i < len(fv); i++ {
		for j < len(r.Vars) && r.Vars[j].Name() < fv[i].Name() {
			j++
		}
		if j < len(r.Vars) && r.Vars[j].Name() != fv[i].Name() { // overshoot
			ffv = append(ffv, fv[i])
		} else if i >= len(r.Vars) { // remaining
			ffv = append(ffv, fv[i])
		}
	}
	ffv = append(ffv, r.Chan.FreeVars()...)
	sort.Slice(ffv, names(ffv).Less)
	return remDup(ffv)
}

// Calculi returns the calculi representation.
func (r *Recv) Calculi() string {
	buf := new(bytes.Buffer)
	t := template.Must(template.New("send").Parse(`
{{- .Chan.Name -}}(
{{- range $i, $v := .Vars -}}
{{- if $i -}},{{- end -}}{{- $v.Name -}}
{{- end -}}).{{ .Cont.Calculi }}`))
	err := t.Execute(buf, r)
	if err != nil {
		log.Println("Executing template:", err)
		return ""
	}
	return buf.String()
}

func (r *Recv) String() string {
	return fmt.Sprintf("Recv(%s, %s)\n%s", r.Chan.Name(), r.Vars, r.Cont)
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

// Calculi returns the calculi representation.
func (r *Repeat) Calculi() string {
	buf := new(bytes.Buffer)
	t := template.Must(template.New("rep").Parse(`!{{- .Proc.Calculi -}}`))
	err := t.Execute(buf, r)
	if err != nil {
		log.Println("Executing template:", err)
		return ""
	}
	return buf.String()
}

func (r *Repeat) String() string {
	return fmt.Sprintf("repeat {\n%s}\n", r.Proc)
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
		if n.Name() == r.Name.Name() {
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

// Calculi returns the calculi representation.
func (r *Restrict) Calculi() string {
	buf := new(bytes.Buffer)
	if _, ok := r.Proc.(*Par); ok {
		t := template.Must(template.New("res").Parse(`(new {{ .Name.Name -}})(
{{- .Proc.Calculi -}})`))
		err := t.Execute(buf, r)
		if err != nil {
			log.Println("Executing template:", err)
			return ""
		}
	} else {
		t := template.Must(template.New("res").Parse(`(new {{ .Name.Name -}})
{{- .Proc.Calculi -}}`))
		err := t.Execute(buf, r)
		if err != nil {
			log.Println("Executing template:", err)
			return ""
		}
	}
	return buf.String()
}

func (r *Restrict) String() string {
	return fmt.Sprintf("scope %s {\n%s}\n", r.Name, r.Proc)
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
	fn = append(fn, s.Chan.FreeNames()...)
	for _, v := range s.Vals {
		fn = append(fn, v.FreeNames()...)
	}
	sort.Slice(fn, names(fn).Less)
	return remDup(fn)
}

// FreeVars of Send is the Vals.
func (s *Send) FreeVars() []Name {
	var fv []Name
	fv = append(fv, s.Chan.FreeVars()...)
	for _, v := range s.Vals {
		fv = append(fv, v.FreeVars()...)
	}
	sort.Slice(fv, names(fv).Less)
	return remDup(fv)
}

// Calculi returns the calculi representation.
func (s *Send) Calculi() string {
	buf := new(bytes.Buffer)
	t := template.Must(template.New("send").Parse(`
{{- .Chan.Name -}}<
{{- range $i, $v := .Vals -}}
{{- if $i -}},{{- end -}}{{- $v.Name -}}
{{- end -}}>`))
	err := t.Execute(buf, s)
	if err != nil {
		log.Println("Executing template:", err)
		return ""
	}
	return buf.String()
}

func (s *Send) String() string {
	return fmt.Sprintf("Send(%s, %s)\n", s.Chan.Name(), s.Vals)
}
