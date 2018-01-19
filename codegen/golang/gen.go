// Copyright 2018 Nicholas Ng
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

// Package golang provides a golang codegen backend for the asyncpi package.
package golang

import (
	"bytes"
	"fmt"
	"io"

	"go.nickng.io/asyncpi"
	"go.nickng.io/asyncpi/types"
)

// Generate writes Go code of the Process to w.
func Generate(p asyncpi.Process, w io.Writer) error {
	asyncpi.Bind(p)
	types.Infer(p)
	if err := types.Unify(p); err != nil {
		return err
	}
	if err := gen(p, w); err != nil {
		return err
	}
	return nil
}

func gen(p asyncpi.Process, w io.Writer) error {
	switch p := p.(type) {
	case *asyncpi.NilProcess:
		w.Write([]byte("/* end */"))
		return nil
	case *asyncpi.Par:
		for i := 0; i < len(p.Procs)-1; i++ {
			w.Write([]byte("go func(){ "))
			if err := gen(p.Procs[i], w); err != nil {
				return err
			}
			w.Write([]byte(" }()\n"))
		}
		if err := gen(p.Procs[len(p.Procs)-1], w); err != nil {
			return err
		}
		return nil
	case *asyncpi.Repeat:
		w.Write([]byte("for { "))
		if err := gen(p.Proc, w); err != nil {
			return err
		}
		w.Write([]byte(" };"))
		return nil
	case *asyncpi.Restrict:
		if chType, ok := p.Name.(types.TypedName).Type().(*types.Chan); ok { // channel is treated differently.
			w.Write([]byte(fmt.Sprintf("%s := make(%s); ", p.Name.(types.TypedName).Name(), chType.String())))
			if err := gen(p.Proc, w); err != nil {
				return err
			}
			return nil
		}
		w.Write([]byte(fmt.Sprintf("var %s %s; ", p.Name.Name(), p.Name.(types.TypedName).Type())))
		if err := gen(p.Proc, w); err != nil {
			return err
		}
		return nil
	case *asyncpi.Recv:
		var buf bytes.Buffer
		switch len(p.Vars) {
		case 0:
			buf.WriteString(fmt.Sprintf("<-%s;", p.Chan.Name()))
		case 1:
			buf.WriteString(fmt.Sprintf("%s := <-%s;", p.Vars[0].Name(), p.Chan.Name()))
		default:
			buf.WriteString(fmt.Sprintf("rcvd := <-%s;", p.Chan.Name()))
			for i, v := range p.Vars {
				if i != 0 {
					buf.WriteRune(',')
				}
				buf.WriteString(fmt.Sprintf("%s", v.Name()))
			}
			buf.WriteString(":=")
			for i := 0; i < len(p.Vars); i++ {
				if i != 0 {
					buf.WriteRune(',')
				}
				buf.WriteString(fmt.Sprintf("rcvd.e%d", i))
			}
			buf.WriteRune(';')
		}
		w.Write(buf.Bytes())
		if err := gen(p.Cont, w); err != nil {
			return err
		}
		return nil
	case *asyncpi.Send:
		var buf bytes.Buffer
		switch len(p.Vals) {
		case 0:
			buf.WriteString(fmt.Sprintf("%s <- struct{}{};", p.Chan.Name()))
		case 1:
			buf.WriteString(fmt.Sprintf("%s <- %s;", p.Chan.Name(), p.Vals[0].Name()))
		default:
			buf.WriteString(fmt.Sprintf("%s <- struct {", p.Chan.Name()))
			for i := 0; i < len(p.Vals); i++ {
				if i != 0 {
					buf.WriteRune(';')
				}
				buf.WriteString(fmt.Sprintf("e%d %s", i, p.Vals[i].(types.TypedName).Type()))
			}
			buf.WriteString(fmt.Sprintf("}{"))
			for i, v := range p.Vals {
				if i != 0 {
					buf.WriteRune(',')
				}
				buf.WriteString(v.Name())
			}
			buf.WriteString(fmt.Sprintf("}"))
		}
		w.Write(buf.Bytes())
		return nil
	}
	return nil
}
