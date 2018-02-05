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

// Command asyncpi is a REPL frontend for the asyncpi package.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/fatih/color"
	"go.nickng.io/asyncpi"
)

var (
	flagColour bool
)

// Command is an interface of a runnable command.
type Command interface {
	// Desc is a short description of the command.
	Desc() string

	// Run is the function to call to execute the command.
	Run()
}

// REPL is the base struct for a REPL-loop.
// Use Prompt() method to start running the REPL loop.
type REPL struct {
	Cmd         map[string]Command
	Interrupted chan os.Signal
	Done        chan error
	hist        []asyncpi.Process

	in  io.Reader
	out io.Writer
	err io.Writer
}

func NewREPL() *REPL {
	r := REPL{
		Done: make(chan error),
		in:   os.Stdin,
		out:  os.Stdout,
		err:  os.Stderr,
	}
	r.Cmd = map[string]Command{
		"help":    &helpCmd{r: &r},
		"exit":    &exitCmd{r: &r},
		"history": &histCmd{r: &r},
		"parse":   &parseCmd{r: &r},
		"reduce":  &reduceCmd{r: &r},
		"show":    &subprocCmd{r: &r},
		"codegen": &codegenCmd{r: &r},
	}
	return &r
}

func (r *REPL) appendHistory(p asyncpi.Process) {
	r.hist = append(r.hist, p)
}

func (r *REPL) replaceHistory(p asyncpi.Process) {
	r.hist[len(r.hist)-1] = p
}

const (
	PromptInit = "async-π> "
	PromptMore = ".......> "
)

var (
	outFprintf = color.New(color.FgCyan).FprintfFunc()
	errFprintf = color.New(color.FgRed).FprintfFunc()
)

func (r *REPL) Responsef(fmt string, a ...interface{}) {
	outFprintf(r.out, fmt, a...)
}

func (r *REPL) Errorf(fmt string, a ...interface{}) {
	errFprintf(r.err, fmt, a...)
}

func (r *REPL) Prompt() {
	for {
		var command string
		fmt.Fprint(r.out, PromptInit)
		select {
		case <-r.Interrupted:
			r.Cmd["exit"].Run()
		default:
			// read first space-delimited string (the command).
			_, err := fmt.Fscanf(r.in, "%s", &command)
			if err != nil {
				if err == io.EOF {
					command = "exit"
				}
			}
			command = strings.TrimSpace(command)
			if cmd, ok := r.Cmd[command]; ok {
				cmd.Run()
			} else {
				r.Errorf("Unrecognised command: %s\n", command)
				r.Cmd["help"].Run()
			}
		}
	}
}

func (r *REPL) Usage() string {
	var buf bytes.Buffer
	buf.WriteString("Commands available:\n")
	for name, cmd := range r.Cmd {
		buf.WriteString(fmt.Sprintf("\t%s\t%s\n", name, cmd.Desc()))
	}
	return buf.String()
}

type helpCmd struct {
	r *REPL
}

func (cmd *helpCmd) Desc() string { return "Display this help message." }
func (cmd *helpCmd) Run()         { cmd.r.Responsef(cmd.r.Usage()) }

type exitCmd struct {
	r *REPL
}

func (cmd *exitCmd) Desc() string { return "Exit." }
func (cmd *exitCmd) Run()         { close(cmd.r.Done) }

type histCmd struct {
	r *REPL
}

func (cmd *histCmd) Desc() string { return "Display history." }
func (cmd *histCmd) Run() {
	if len(cmd.r.hist) == 0 {
		cmd.r.Responsef("History is empty.\n")
	}
	for i, p := range cmd.r.hist {
		cmd.r.Responsef("%d:\t%s\n", i, p.Calculi())
	}
}

func init() {
	flag.BoolVar(&flagColour, "colour", true, "Output with colour (needs ANSI colour support)")
}

func main() {
	flag.Parse()
	color.NoColor = !flagColour
	repl := NewREPL()
	repl.Interrupted = make(chan os.Signal, 1)
	signal.Notify(repl.Interrupted, os.Interrupt)
	go repl.Prompt()
	err := <-repl.Done
	if err != nil {
		repl.Errorf("asyncpi error: %v\n", err)
		os.Exit(1)
	}
}
