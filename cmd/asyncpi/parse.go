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
	"bufio"
	"bytes"
	"fmt"
	"io"

	"go.nickng.io/asyncpi"
)

type parseCmd struct {
	r *REPL
}

func (cmd *parseCmd) Desc() string {
	return "Parse an asynchronous π-calculus process."
}

func (cmd *parseCmd) Run() {
	cmd.r.Responsef("%s\n", cmd.parse(cmd.r.in))
}

func (cmd *parseCmd) parse(r io.Reader) string {
	br := bufio.NewReader(r)
	var err error
	var b []byte
	var buf bytes.Buffer
	for err != io.EOF {
		if err != io.EOF {
			fmt.Fprint(cmd.r.out, PromptMore)
		}
		b, err = br.ReadBytes('\n')
		buf.Write(b)
	}
	fmt.Fprintln(cmd.r.out)
	var cached bytes.Buffer
	proc, err := asyncpi.Parse(io.TeeReader(&buf, &cached))
	if err != nil {
		if parseErr, ok := err.(*asyncpi.ParseError); ok {
			cmd.r.Errorf("Parse failed:\n%s", string(parseErr.Pos.CaretDiag(cached.Bytes())))
			return ""
		}
		return ""
	}
	cmd.r.appendHistory(proc)
	return proc.Calculi()
}
