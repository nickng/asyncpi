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

package main

import (
	"bytes"
	"go/format"

	"go.nickng.io/asyncpi"
	"go.nickng.io/asyncpi/codegen/golang"
	"go.nickng.io/asyncpi/types"
)

type codegenCmd struct {
	r *REPL
}

func (cmd *codegenCmd) Desc() string {
	return "Generate a fragment of Go code."
}

func (cmd *codegenCmd) Run() {
	if len(cmd.r.hist) < 1 {
		cmd.r.Errorf("No last process to generate from.\n")
		return
	}
	p := cmd.r.hist[len(cmd.r.hist)-1]
	if err := asyncpi.Bind(&p); err != nil {
		cmd.r.Done <- err
		return
	}
	err := types.Infer(p)
	if err != nil {
		cmd.r.Done <- err
		return
	}
	err = types.Unify(p)
	if err != nil {
		cmd.r.Done <- err
		return
	}
	var output bytes.Buffer
	err = golang.Generate(p, &output)
	if err != nil {
		cmd.r.Done <- err
		return
	}
	b, err := format.Source(output.Bytes())
	if err != nil {
		cmd.r.Done <- err
		return
	}
	cmd.r.Responsef("/* start generated code */\n\n")
	cmd.r.Responsef("%s\n\n", string(b))
	cmd.r.Responsef("/* end generated code */\n")
}
