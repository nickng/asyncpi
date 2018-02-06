package asyncpi

// Binding.
// This file contains functions for name binding.

// Bind takes a parsed process p and returned a process with valid binding.
func Bind(p *Process) error {
	var err error
	*p, err = bind(*p, []Name{})
	if err != nil {
		return err
	}
	return nil
}

// bind is a depth-first recursive traversal of a Process p
// with boundNames to bind Names with the same Ident.
func bind(p Process, boundNames []Name) (_ Process, err error) {
	switch p := p.(type) {
	case *NilProcess:
		return p, nil
	case *Repeat:
		p.Proc, err = bind(p.Proc, boundNames)
		return p, err
	case *Par:
		for i := range p.Procs {
			p.Procs[i], err = bind(p.Procs[i], boundNames)
		}
		return p, err
	case *Recv:
		names := make([]Name, len(boundNames))
		for i := range boundNames {
			names[i] = boundNames[i]
		}
		names = append(names, p.Vars...)
		for _, v := range p.Vars {
			for j := 0; j < len(names)-len(p.Vars); j++ {
				if IsSameName(v, names[j]) {
					// Rebinding existing bound name.
					names = append(names[:j], names[j+1:]...)
				}
			}
		}
		var chanBound bool
		for i, bn := range names {
			if IsSameName(p.Chan, bn) { // Found bound Chan
				p.Chan = names[i]
				chanBound = true
			}
		}
		if !chanBound {
			names = append(names, p.Chan)
		}
		p.Cont, err = bind(p.Cont, names)
		return p, err
	case *Send:
		count := 0
		for i, bn := range boundNames {
			for j, v := range p.Vals {
				if IsSameName(v, bn) { // Found bound name.
					p.Vals[j] = boundNames[i]
					count++
				}
			}
		}
		for i, bn := range boundNames {
			if IsSameName(p.Chan, bn) { // Found bound Chan.
				p.Chan = boundNames[i]
				count++
			}
		}
		return p, err
	case *Restrict:
		names := append(boundNames, p.Name)
		for i := 0; i < len(names)-1; i++ {
			if IsSameName(p.Name, names[i]) {
				// Rebinding existing bound name.
				names = append(names[:i], names[i+1:]...)
			}
		}
		p.Proc, err = bind(p.Proc, names)
		return p, err
	default:
		return nil, UnknownProcessTypeError{Caller: "bind", Proc: p}
	}
}
