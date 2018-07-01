// Copyright 2018 Google Inc.
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

package model

import (
	"bytes"
	"encoding/json"
	"go/format"
	"io"
	"text/template"

	"github.com/google/shenzhen-go/dev/source"
)

const (
	// Fun fact: go/format fixes trailing commas in function args.
	goTemplateSrc = `{{if .IsCommand}}
// The {{.PackageName}} command was automatically generated by Shenzhen Go.
package main
{{else}}
// Package {{.PackageName}} was automatically generated by Shenzhen Go.
package {{.PackageName}}{{if ne .PackagePath .PackageName}} // import "{{.PackagePath}}"{{end}}
{{end}}

import (
	{{range .AllImports}}
	{{.}}
	{{- end}}
)

{{range .Nodes}}
func {{.Identifier}}({{range $name, $type := .PinFullTypes}}{{$name}} {{$type}},
	{{end}}) {
	{{.ImplHead}}
	{{if eq .Multiplicity 1 -}}
	func(instanceNumber, multiplicity int) {
		{{.ImplBody}}
	}(0, 1)
	{{- else -}}
	var multWG sync.WaitGroup
	multWG.Add({{.Multiplicity}})
	for n:=0; n<{{.Multiplicity}}; n++ {
		go func(instanceNumber, multiplicity int) {
			defer multWG.Done()
			{{.ImplBody}}
		}(n, {{.Multiplicity}})
	}
	multWG.Wait()
	{{- end}}
	{{.ImplTail}}
}
{{end}}

{{if .IsCommand}}
func main() {
{{else}}
// Run executes all the goroutines associated with the graph that generated 
// this package, and waits for any that were marked as "wait for this to 
// finish" to finish before returning.
func Run() {
{{end}}
	{{- range $n, $c := .Channels}}
	{{$n}} := make(chan {{$c.Type}}, {{$c.Capacity}})
	{{- end}}

	var wg sync.WaitGroup
	{{range $node := .Nodes}}
		{{if $node.Enabled -}}
			{{if $node.Wait -}}
	wg.Add(1)
	go func() {
			{{$node.Identifier}}({{range $pin := $node.Pins}}{{index $node.Connections $pin.Name}},{{end}})
		wg.Done()
	}()
			{{else}}
	go {{$node.Identifier}}({{range $pin := $node.Pins}}{{index $node.Connections $pin.Name}},{{end}})
			{{- end}}
		{{- end}}
	{{- end}}

	// Wait for the end
	wg.Wait()
}`
)

var goTemplate = template.Must(template.New("golang").Parse(goTemplateSrc))

// WriteRawGoTo writes the Go language view of the graph to the io.Writer, without gofmt-ing.
func (g *Graph) WriteRawGoTo(w io.Writer) error {
	if err := g.InferTypes(); err != nil {
		return err
	}
	return goTemplate.Execute(w, g)
}

// RawGo outputs the unformatted Go language view of the graph.
func (g *Graph) RawGo() (string, error) {
	buf := &bytes.Buffer{}
	if err := g.WriteRawGoTo(buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// WriteGoTo writes the Go language view of the graph to the io.Writer.
func (g *Graph) WriteGoTo(w io.Writer) error {
	buf := &bytes.Buffer{}
	if err := g.WriteRawGoTo(buf); err != nil {
		return err
	}
	return source.GoFmt(w, buf)
}

// Go outputs the Go language view of the graph.
func (g *Graph) Go() (string, error) {
	buf := &bytes.Buffer{}
	if err := g.WriteRawGoTo(buf); err != nil {
		return "", err
	}
	o, err := format.Source(buf.Bytes())
	return string(o), err
}

// WriteJSONTo writes nicely-formatted JSON to the given Writer.
func (g *Graph) WriteJSONTo(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t") // For diffability!
	return enc.Encode(g)
}

// JSON returns the JSON view of the graph.
func (g *Graph) JSON() (string, error) {
	o, err := json.MarshalIndent(g, "", "\t")
	return string(o), err
}
