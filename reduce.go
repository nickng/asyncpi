package asyncpi

import (
	"fmt"
)

// Subst is the substitution of variables xs by names vs in Process p.
//
// The caller must ensure that the size of xs and vs are the same
// such that xs[i] is substituted by vs[i] in 0 <= i < len(xs).
func Subst(p Process, vs, xs []Name) error {
	if len(xs) != len(vs) {
		return ErrInvalid
	}
	procs := []Process{p}
	for len(procs) > 0 {
		p, procs = procs[0], procs[1:]
		switch p := p.(type) {
		case *NilProcess:
		case *Par:
			procs = append(procs, p.Procs...)
		case *Recv:
			for i, x := range xs {
				if IsSameName(p.Chan, x) {
					if ch, canSetName := p.Chan.(nameSetter); canSetName {
						ch.setName(vs[i].Name())
					}
				}
				for _, rv := range p.Vars {
					if IsSameName(rv, x) {
						if ch, canSetName := rv.(nameSetter); canSetName {
							ch.setName(vs[i].Name())
						}
					}
				}
			}
			procs = append(procs, p.Cont)
		case *Restrict:
			procs = append(procs, p.Proc)
		case *Repeat:
			procs = append(procs, p.Proc)
		case *Send:
			for i, x := range xs {
				if IsSameName(p.Chan, x) {
					if ch, canSetName := p.Chan.(nameSetter); canSetName {
						ch.setName(vs[i].Name())
					}
				}
			}
		default:
			return UnknownProcessTypeError{Caller: "Subst", Proc: p}
		}
	}
	return nil
}

// reduceOnce performs a single step of reduction.
//
// This reduction combines multiple congruence relations
// and attempts to perform an interaction.
func reduceOnce(p Process) (changed bool, err error) {
	switch p := p.(type) {
	case *NilProcess:
		return false, nil
	case *Par:
		var sends map[string]*Process // pointer because we will mutate them
		var recvs map[string]*Process // pointer because we will mutate them
		for i, proc := range p.Procs {
			switch proc := proc.(type) {
			case *Par:
				// nested Par.
				return reduceOnce(proc)
			case *Recv:
				if recvs == nil {
					recvs = make(map[string]*Process)
				}
				if IsFreeName(proc.Chan) {
					ch := proc.Chan.Name()
					// Do not overwrite existing subprocess with same channel.
					// Substitution only consider leftmost available names.
					if _, exists := recvs[ch]; !exists {
						recvs[ch] = &p.Procs[i]
					}
				}
			case *Send:
				if sends == nil {
					sends = make(map[string]*Process)
				}
				if IsFreeName(proc.Chan) {
					ch := proc.Chan.Name()
					// Do not overwrite existing subprocess with same channel.
					// Substitution only consider leftmost available names.
					if _, exists := sends[ch]; !exists {
						sends[ch] = &p.Procs[i]
					}
				}
			}
		}
		for ch, s := range sends {
			if r, hasSharedChan := recvs[ch]; hasSharedChan {
				recv := (*r).(*Recv)
				send := (*s).(*Send)
				if err := Subst(recv.Cont, send.Vals, recv.Vars); err != nil {
					return false, fmt.Errorf("substitution error: %v", err)
				}
				*s, *r = NewNilProcess(), recv.Cont
				return true, nil
			}
		}
		return false, nil
	case *Recv:
		return false, nil
	case *Restrict:
		return reduceOnce(p.Proc)
	case *Repeat:
		return reduceOnce(p.Proc)
	case *Send:
		return false, nil
	default:
		return false, UnknownProcessTypeError{Caller: "reduceOnce", Proc: p}
	}
}

// SimplifyBySC simplifies a Process p by structural congruence rules.
//
// It applies functions to (1) remove unnecessary Restrict,
// and (2) remove superfluous inact.
//
func SimplifyBySC(p Process) (Process, error) {
	unwanted, err := findUnusedRestrict(p)
	if err != nil {
		return nil, err
	}
	p, err = filterRestrict(p, unwanted)
	if err != nil {
		return nil, err
	}
	p, err = filterNilProcess(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// findUnusedRestrict returns a slice of unused Names
// introduced by a Restrict in Process p.
//
// More accurately, this function returns all x in the form of:
//
//     (νx)P where x ∉ fn(P)
//
// Such that the unused Names and the Restrict that
// introduced them can be removed.
func findUnusedRestrict(p Process) (unused []Name, err error) {
	p = Bind(p)
	type rescount struct {
		ResName Name
		Count   int
	}
	resUses := make(map[string]*rescount)
	procs := []Process{p}
	for len(procs) > 0 {
		p, procs = procs[0], procs[1:]
		switch p := p.(type) {
		case *NilProcess:
		case *Par:
			procs = append(procs, p.Procs...)
		case *Recv:
			if rc, exists := resUses[p.Chan.Name()]; exists {
				rc.Count++
			}
			for _, v := range p.Vars {
				if rc, exists := resUses[v.Name()]; exists {
					rc.Count++
				}
			}
			procs = append(procs, p.Cont)
		case *Repeat:
			procs = append(procs, p.Proc)
		case *Restrict:
			resUses[p.Name.Name()] = &rescount{ResName: p.Name, Count: 1}
			procs = append(procs, p.Proc)
		case *Send:
			if rc, exists := resUses[p.Chan.Name()]; exists {
				rc.Count++
			}
			for _, v := range p.Vals {
				if rc, exists := resUses[v.Name()]; exists {
					rc.Count++

				}
			}
		default:
			return nil, UnknownProcessTypeError{Caller: "findUnusedRestrict", Proc: p}
		}
	}
	for _, rc := range resUses {
		if rc.Count == 1 {
			unused = append(unused, rc.ResName)
		}
	}
	return unused, nil
}

// filterRestrict returns a Process where all unwanted Names
// introduced by Restrict are removed from the input Process p.
//
// More accurately, this function returns P given a slice of x:
//
//     (νx)P where x ∉ fn(P)
//
// See also findUnusedRestrict(Process) function for creating the
// unwanted Name slice.
func filterRestrict(p Process, unwanted []Name) (Process, error) {
	switch p := p.(type) {
	case *NilProcess:
		return p, nil
	case *Par:
		var procs []Process
		for _, proc := range p.Procs {
			res, err := filterRestrict(proc, unwanted)
			if err != nil {
				return nil, err
			}
			procs = append(procs, res)
		}
		p.Procs = procs
		return p, nil
	case *Recv:
		var err error
		p.Cont, err = filterRestrict(p.Cont, unwanted)
		if err != nil {
			return nil, err
		}
		return p, nil
	case *Repeat:
		var err error
		p.Proc, err = filterRestrict(p.Proc, unwanted)
		if err != nil {
			return nil, err
		}
		return p, nil
	case *Restrict:
		var err error
		p.Proc, err = filterRestrict(p.Proc, unwanted)
		if err != nil {
			return nil, err
		}
		for _, n := range unwanted {
			if n == p.Name { // note: pointer compare
				return p.Proc, nil
			}
		}
		return p, nil
	case *Send:
		return p, nil
	default:
		return nil, UnknownProcessTypeError{Caller: "filterRestrict", Proc: p}
	}
}

// filterNilProcess returns a Process where superfluous NilProcess
// that does not end a Process are removed from the input Process p.
//
// More accurately, this function removes 0 or collapses Process involving 0:
//
//     (P|0) → P
//     !0 → 0
//
// NilProcess at the end of Processes are unchanged.
func filterNilProcess(p Process) (Process, error) {
	switch p := p.(type) {
	case *NilProcess:
		return p, nil
	case *Par:
		var procs []Process
		for _, proc := range p.Procs {
			proc, err := filterNilProcess(proc)
			if err != nil {
				return nil, err
			}
			if _, isEmpty := proc.(*NilProcess); !isEmpty {
				procs = append(procs, proc)
			}
		}
		switch len(procs) {
		case 0:
			return NewNilProcess(), nil
		case 1:
			return procs[0], nil
		default:
			p.Procs = procs
			return p, nil
		}
	case *Recv:
		var err error
		p.Cont, err = filterNilProcess(p.Cont)
		if err != nil {
			return nil, err
		}
		return p, nil
	case *Repeat:
		var err error
		p.Proc, err = filterNilProcess(p.Proc)
		if err != nil {
			return nil, err
		}
		if _, isEmpty := p.Proc.(*NilProcess); isEmpty {
			return p.Proc, nil
		}
		return p, nil
	case *Restrict:
		var err error
		p.Proc, err = filterNilProcess(p.Proc)
		if err != nil {
			return nil, err
		}
		if _, isEmpty := p.Proc.(*NilProcess); isEmpty {
			return p.Proc, nil
		}
		return p, nil
	case *Send:
		return p, nil
	default:
		return nil, UnknownProcessTypeError{Caller: "filterNilProcess", Proc: p}
	}
}
