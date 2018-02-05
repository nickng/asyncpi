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
	"go.nickng.io/asyncpi"
)

type reduceCmd struct {
	r *REPL
}

func (cmd *reduceCmd) Desc() string {
	return "Reduce the last parsed process."
}

func (cmd *reduceCmd) Run() {
	if len(cmd.r.hist) < 1 {
		cmd.r.Errorf("No last process to reduce from.\n")
		return
	}
	p := cmd.r.hist[len(cmd.r.hist)-1]
	cmd.r.Responsef("Reducing: %s\n", p.Calculi())
	cmd.reduce(p)
}

func (cmd *reduceCmd) reduce(p asyncpi.Process) {
	if err := asyncpi.Bind(&p); err != nil {
		cmd.r.Done <- err
		return
	}
	changed, err := asyncpi.Reduce1(p)
	if err != nil {
		cmd.r.Done <- err
		return
	}
	if changed {
		p, err = asyncpi.SimplifyBySC(p)
		if err != nil {
			cmd.r.Done <- err
			return
		}
	}
	cmd.r.Responsef("%s\n", p.Calculi())
	cmd.r.replaceHistory(p)
}
