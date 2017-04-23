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
	Type() Type
	SetType(t Type)
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

type sorts int

const (
	nameSort sorts = iota // default sort is name.
	varSort
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

// Process is process prefixed with action.
type Process interface {
	FreeNames() []Name
	FreeVars() []Name

	Calculi() string
	Golang() string

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

// Golang returns the Go representation.
func (n *NilProcess) Golang() string {
	return "/* end */"
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

// Golang returns the Golang representation.
func (p *Par) Golang() string {
	var buf bytes.Buffer
	for i := 0; i < len(p.Procs)-1; i++ {
		buf.WriteString(fmt.Sprintf("go func(){ %s }()\n", p.Procs[i].Golang()))
	}
	buf.WriteString(p.Procs[len(p.Procs)-1].Golang())
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

// Golang returns the Go representation.
func (r *Recv) Golang() string {
	var buf bytes.Buffer
	switch len(r.Vars) {
	case 0:
		buf.WriteString(fmt.Sprintf("<-%s;", r.Chan.Name()))
	case 1:
		buf.WriteString(fmt.Sprintf("%s := <-%s;", r.Vars[0].Name(), r.Chan.Name()))
	default:
		buf.WriteString(fmt.Sprintf("rcvd := <-%s;", r.Chan.Name()))
		for i, v := range r.Vars {
			if i != 0 {
				buf.WriteRune(',')
			}
			buf.WriteString(fmt.Sprintf("%s", v.Name()))
		}
		buf.WriteString(":=")
		for i := 0; i < len(r.Vars); i++ {
			if i != 0 {
				buf.WriteRune(',')
			}
			buf.WriteString(fmt.Sprintf("rcvd.e%d", i))
		}
		buf.WriteRune(';')
	}
	buf.WriteString(r.Cont.Golang())
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

// Golang returns the Go representation.
func (r *Repeat) Golang() string {
	return fmt.Sprintf("for { %s };", r.Proc.Golang())
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

// Golang returns the Go representation.
func (r *Restrict) Golang() string {
	if chType, ok := r.Name.Type().(*chanType); ok { // channel is treated differently.
		return fmt.Sprintf("%s := make(%s); %s", r.Name.Name(), chType.String(), r.Proc.Golang())
	}
	return fmt.Sprintf("var %s %s; %s", r.Name.Name(), r.Name.Type(), r.Proc.Golang())
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

// Golang returns the Go representation.
func (s *Send) Golang() string {
	var buf bytes.Buffer
	switch len(s.Vals) {
	case 0:
		buf.WriteString(fmt.Sprintf("%s <- struct{}{};", s.Chan.Name()))
	case 1:
		buf.WriteString(fmt.Sprintf("%s <- %s;", s.Chan.Name(), s.Vals[0].Name()))
	default:
		buf.WriteString(fmt.Sprintf("%s <- struct {", s.Chan.Name()))
		for i := 0; i < len(s.Vals); i++ {
			if i != 0 {
				buf.WriteRune(';')
			}
			buf.WriteString(fmt.Sprintf("e%d %s", i, s.Vals[i].Type()))
		}
		buf.WriteString(fmt.Sprintf("}{"))
		for i, v := range s.Vals {
			if i != 0 {
				buf.WriteRune(',')
			}
			buf.WriteString(v.Name())
		}
		buf.WriteString(fmt.Sprintf("}"))

	}
	return buf.String()
}

func (s *Send) String() string {
	return fmt.Sprintf("Send(%s, %s)\n", s.Chan.Name(), s.Vals)
}

// GenerateGo generates Go code from a toplevel Process proc.
func GenerateGo(proc Process) string {
	proc = Bind(proc)
	Infer(proc)
	err := Unify(proc)
	if err != nil {
		log.Fatal(err)
	}
	return proc.Golang()
}
