// Copyright 2018 Nicholas Ng <nickng@nickng.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"fmt"

	"go.nickng.io/asyncpi"
)

// BUG(nickng) Inference may fail if type of a name is recursively defined
// (e.g. a<a> → typed chan of type(a)), printing the type will cause a stack
// overflow.

func processAttachType(p asyncpi.Process) {
	switch p := p.(type) {
	case *asyncpi.NilProcess:
	case *asyncpi.Par:
		for _, p := range p.Procs {
			processAttachType(p)
		}
	case *asyncpi.Recv:
		p.Chan = AttachType(p.Chan)
		var tvs []asyncpi.Name
		for _, v := range p.Vars {
			tvs = append(tvs, AttachType(v))
		}
		p.SetVars(tvs)
		processAttachType(p.Cont)
	case *asyncpi.Repeat:
		processAttachType(p.Proc)
	case *asyncpi.Restrict:
		processAttachType(p.Proc)
		p.Name = AttachType(p.Name)
	case *asyncpi.Send:
		p.Chan = AttachType(p.Chan)
		var tvs []asyncpi.Name
		for _, v := range p.Vals {
			tvs = append(tvs, AttachType(v))
		}
		p.SetVals(tvs)
	}
}

func Infer(p asyncpi.Process) error {
	processAttachType(p)
	if err := asyncpi.Bind(&p); err != nil {
		return err
	}
	return processInferType(p)
}

// processInferType performs inline type inference for channels.
//
// processInferType should be called after Bind, so the types of names inferred from
// channels can be propagated to other references bound to the same name.
func processInferType(p asyncpi.Process) error {
	switch p := p.(type) {
	case *asyncpi.NilProcess:
	case *asyncpi.Par:
		for _, p := range p.Procs {
			if err := Infer(p); err != nil {
				return err
			}
		}
	case *asyncpi.Recv:
		if err := Infer(p.Cont); err != nil {
			return err
		}
		if _, isTyped := p.Chan.(TypedName); !isTyped {
			return InferUntypedError{Name: p.Chan.Ident()}
		}
		// But that's all we know right now.
		if _, ok := p.Chan.(TypedName).Type().(*anyType); ok { // do not overwrite existing type
			var tvs []Type
			for i := range p.Vars {
				tv, isTyped := p.Vars[i].(TypedName)
				if !isTyped {
					return InferUntypedError{Name: p.Vars[i].Ident()}
				}
				if refType, isRef := tv.Type().(*Reference); isRef { // already a Reference
					tvs = append(tvs, refType)
				} else {
					tvs = append(tvs, NewReference(p.Vars[i]))
				}
			}
			switch len(tvs) { // arity determines if it is a Composite.
			case 1:
				p.Chan.(TypedName).setType(NewChan(tvs[0]))
			default:
				p.Chan.(TypedName).setType(NewChan(NewComposite(tvs...)))
			}
		}
	case *asyncpi.Repeat:
		if err := Infer(p.Proc); err != nil {
			return err
		}
	case *asyncpi.Restrict:
		if err := Infer(p.Proc); err != nil {
			return err
		}
	case *asyncpi.Send: // Send is the only place we can infer channel type.
		if _, isTyped := p.Chan.(TypedName); !isTyped {
			return InferUntypedError{Name: p.Chan.Ident()}
		}
		var tvs []Type
		for i := range p.Vals {
			if _, isTyped := p.Vals[i].(TypedName); !isTyped {
				return InferUntypedError{Name: p.Vals[i].Ident()}
			}
			if refType, isRef := p.Vals[i].(TypedName).Type().(*Reference); isRef { // already a Reference
				tvs = append(tvs, refType)
			} else {
				tvs = append(tvs, NewReference(p.Vals[i]))
			}
		}
		switch len(p.Vals) {
		case 1:
			p.Chan.(TypedName).setType(NewChan(tvs[0]))
		default:
			p.Chan.(TypedName).setType(NewChan(NewComposite(tvs...)))
		}
	default:
		return asyncpi.InvalidProcTypeError{Caller: "types.Infer", Proc: p}
	}
	return nil
}

// Unify combines the constraints of sending and receiving channels
// with best effort.
//
// It is assumed that the names are already typed, and an error is returned
// if the typing constraints are in conflict and cannot be unified.
// A Process is well-typed if no error is returned.
func Unify(p asyncpi.Process) error {
	switch p := p.(type) {
	case *asyncpi.NilProcess, *asyncpi.Send: // No continuation.
	case *asyncpi.Par:
		for _, p := range p.Procs {
			if err := Unify(p); err != nil {
				return err
			}
		}
	case *asyncpi.Recv:
		ch, isTyped := p.Chan.(TypedName)
		if !isTyped {
			return InferUntypedError{Name: p.Chan.Ident()}
		}
		// chType is either
		// - a compType with refType fields (including struct{})
		// - a refType (non-tuple)
		varType := ch.Type().(*Chan).Elem()
		var tns []TypedName
		for _, v := range p.Vars {
			tv, isTyped := v.(TypedName)
			if !isTyped {
				return InferUntypedError{Name: v.Ident()}
			}
			tns = append(tns, tv)
		}
		switch len(p.Vars) {
		case 1:
			refT, isRef := varType.(*Reference)
			if !isRef {
				return &TypeArityError{
					Got:      len(varType.(*Composite).Elems()),
					Expected: 1,
					Msg:      fmt.Sprintf("Types from channel %s and vars have different arity", p.Chan.Ident()),
				}
			}
			if _, ok := tns[0].Type().(*anyType); ok {
				tns[0].setType(varType)
			} else {
				if _, ok := refT.ref.Type().(*anyType); ok {
					refT.ref.setType(tns[0].Type())
				} else if IsEqual(varType, tns[0].Type()) {
					// Type is both set but equal
				} else {
					return &TypeError{
						T:   varType,
						U:   tns[0].Type(),
						Msg: fmt.Sprintf("Types inferred from channel %s are in conflict", p.Chan.Ident()),
					}
				}
			}
		default:
			compT, isComp := varType.(*Composite)
			if !isComp {
				return &TypeArityError{
					Got:      1,
					Expected: len(p.Vars),
					Msg:      fmt.Sprintf("Types from channel %s and vars have different arity", p.Chan.Ident()),
				}
			} else if len(tns) != len(compT.Elems()) {
				return &TypeArityError{
					Got:      len(compT.Elems()),
					Expected: len(p.Vars),
					Msg:      fmt.Sprintf("Types from channel %s and vars have different arity", p.Chan.Ident()),
				}
			}
			for i := range tns {
				if _, ok := tns[i].Type().(*anyType); ok {
					tns[i].setType(compT.elems[i].(*Reference).ref.Type())
				} else if _, ok := compT.elems[i].(*Reference).ref.Type().(*anyType); ok {
					compT.elems[i].(*Reference).ref.setType(tns[i].Type())
				} else if IsEqual(compT.elems[i], tns[i].Type()) {
					// Type is both set but equal
				} else {
					return &TypeError{
						T:   varType,
						U:   tns[i].Type(),
						Msg: fmt.Sprintf("Types inferred from channel %s are in conflict", p.Chan.Ident()),
					}
				}
			}
		}
		return Unify(p.Cont)
	case *asyncpi.Repeat:
		return Unify(p.Proc)
	case *asyncpi.Restrict:
		return Unify(p.Proc)
	}
	return nil
}
