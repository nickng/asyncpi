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

package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"io"

	"go.nickng.io/asyncpi"
	"golang.org/x/tools/imports"
)

// FormatOptions defines options for changing
// the format of generated code.
type FormatOptions struct {
	Main     bool
	Debug    bool
	Format   bool
	FmtStyle FormatStyle
}

// FormatStyle defines the tools to use for formatting code
type FormatStyle int

const (
	// Gofmt is the option to use go/format style
	Gofmt FormatStyle = iota
	// GoImports is the option to use goimports style
	GoImports
)

// GenerateOpts writes Go code of the Process p to w using options opt.
func GenerateOpts(p asyncpi.Process, opt FormatOptions, w io.Writer) error {
	var program bytes.Buffer
	const progHeader = `package main
	func main() {
`
	const progFooter = `
}`

	if opt.Main {
		fmt.Fprintf(&program, progHeader)
	}
	if opt.Debug {
		fmt.Fprintf(&program, "// Process %s\n", p.Calculi())
		fmt.Fprint(&program, `fmt.Fprintln(os.Stderr, "--- start ---");`)
	}
	if err := Generate(p, &program); err != nil {
		return err
	}
	if opt.Debug {
		fmt.Fprint(&program, `fmt.Fprintln(os.Stderr, "--- end ---");`)
	}
	if opt.Main {
		fmt.Fprint(&program, progFooter)
	}
	if opt.Format {
		switch opt.FmtStyle {
		case Gofmt:
			b, err := format.Source(program.Bytes())
			if err != nil {
				return err
			}
			_, err = w.Write(b)
			if err != nil {
				return err
			}
		case GoImports:
			b, err := imports.Process("/tmp/tmp.go", program.Bytes(), &imports.Options{
				Comments:  true,
				Fragment:  !opt.Main,
				TabIndent: true,
			})
			if err != nil {
				return err
			}
			_, err = w.Write(b)
			if err != nil {
				return err
			}
		}
		return nil
	}
	_, err := io.Copy(w, &program)
	if err != nil {
		return err
	}
	return nil
}
