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

package asyncpi

import (
	"bytes"
	"log"
	"text/template"
)

const parTmpl = `(
{{- range $i, $p := .Procs -}}
{{- if $i }} | {{ end -}}{{- $p.Calculi -}}
{{- end -}})`

const recvTmpl = `{{- .Chan.Ident -}}(
{{- range $i, $v := .Vars -}}
{{- if $i -}},{{- end -}}{{- $v.Ident -}}
{{- end -}}).{{ .Cont.Calculi }}`

const sendTmpl = `{{- .Chan.Ident -}}<
{{- range $i, $v := .Vals -}}
{{- if $i -}},{{- end -}}{{- $v.Ident -}}
{{- end -}}>`

const repTmpl = `!{{- .Proc.Calculi -}}`

const resTmpl = `(new {{ .Name.Ident -}}){{- .Proc.Calculi -}}`

var (
	parT  = template.Must(template.New("").Parse(parTmpl))
	recvT = template.Must(template.New("").Parse(recvTmpl))
	repT  = template.Must(template.New("").Parse(repTmpl))
	resT  = template.Must(template.New("").Parse(resTmpl))
	sendT = template.Must(template.New("").Parse(sendTmpl))
)

func (p *NilProcess) Calculi() string {
	return "0"
}

func (p *Par) Calculi() string {
	var buf bytes.Buffer
	if err := parT.Execute(&buf, p); err != nil {
		log.Print(err)
	}
	return buf.String()
}

func (p *Recv) Calculi() string {
	var buf bytes.Buffer
	if err := recvT.Execute(&buf, p); err != nil {
		log.Print(err)
	}
	return buf.String()
}

func (p *Repeat) Calculi() string {
	var buf bytes.Buffer
	if err := repT.Execute(&buf, p); err != nil {
		log.Print(err)
	}
	return buf.String()
}

func (p *Restrict) Calculi() string {
	var buf bytes.Buffer
	if err := resT.Execute(&buf, p); err != nil {
		log.Print(err)
	}
	return buf.String()
}

func (p *Send) Calculi() string {
	var buf bytes.Buffer
	if err := sendT.Execute(&buf, p); err != nil {
		log.Print(err)
	}
	return buf.String()
}
