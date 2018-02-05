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
	"fmt"

	"go.nickng.io/asyncpi"
	"go.nickng.io/asyncpi/name/sortedname"
)

type subprocCmd struct {
	r *REPL
}

func (cmd *subprocCmd) Desc() string {
	return "Display subprocesses of the last parsed process."
}

func (cmd *subprocCmd) Run() {
	if len(cmd.r.hist) < 1 {
		cmd.r.Errorf("No last process to show.\n")
		return
	}
	p := cmd.r.hist[len(cmd.r.hist)-1]
	if err := sortedname.InferSortsByUsage(p); err != nil {
		cmd.r.Done <- err
		return
	}
	cmd.displaySubprocess(p)
}

func (cmd *subprocCmd) displaySubprocess(p asyncpi.Process) {
	procs := []asyncpi.Process{p}
	for len(procs) > 0 {
		p, procs = procs[0], procs[1:]
		cmd.r.Responsef("%s\n\tfn = %q\n\tfv = %q\n", p.Calculi(), p.FreeNames(), p.FreeVars())
		switch p := p.(type) {
		case *asyncpi.NilProcess:
		case *asyncpi.Par:
			procs = append(procs, p.Procs...)
		case *asyncpi.Recv:
			procs = append(procs, p.Cont)
		case *asyncpi.Repeat:
			procs = append(procs, p.Proc)
		case *asyncpi.Restrict:
			procs = append(procs, p.Proc)
		case *asyncpi.Send:
		default:
			cmd.r.Done <- fmt.Errorf("unknown subprocess type: %s", p.Calculi())
			return
		}
	}
}
