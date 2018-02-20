// Copyright 2017 Google Inc.
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

package parts

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/google/shenzhen-go/dev/source"
)

const unslicerEditorTemplateSrc = `
<div class="formfield">
	<label for="Input">Input</label>
	<select name="Input">
		{{range .Graph.Channels -}}
		<option value="{{.Name}}" {{if eq .Name $.Node.Part.Input}}selected{{end}}>{{.Name}}</option>
		{{- end}}
	</select>
</div>
<div class="formfield">
	<label for="Output">Output</label>
	<select name="Output">
		{{range .Graph.Channels -}}
		<option value="{{.Name}}" {{if eq .Name $.Node.Part.Output}}selected{{end}}>{{.Name}}</option>
		{{- end}}
	</select>
</div>
`

// Unslicer ranges over items that arrive via the input, and sends them
// individually to the output.
type Unslicer struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// AssociateEditor associates a template called "part_view" with the given template.
func (*Unslicer) AssociateEditor(t *template.Template) error {
	_, err := t.New("part_view").Parse(unslicerEditorTemplateSrc)
	return err
}

// Channels returns any channels used. Anything returned that is not a channel is ignored.
func (p *Unslicer) Channels() (read, written source.StringSet) {
	return source.NewStringSet(p.Input), source.NewStringSet(p.Output)
}

// Clone returns a copy of this part.
func (p *Unslicer) Clone() interface{} {
	s := *p
	return &s
}

// Help returns useful help information.
func (*Unslicer) Help() template.HTML {
	return `<p>
	The Unslicer reads slices from the input channel, and sends the elements individually to the output.
	When the input is closed, the output is closed after all remaining slice elements to send have been sent.
	</p><p>
	For example, if the input receives:
	<p><code>[]int{2, 3, 5}<br/>[]int{1, 4, 6}</code></p>
	then this sequence of integers will be sent to the output:
	<p><code>2, 3, 5, 1, 4, 6</code></p>
	</p>
	`
}

// Impl returns Go source code implementing the part.
func (p *Unslicer) Impl() (head, body, tail string) {
	return "",
		fmt.Sprintf(`for i := range %s { for _, o := range i { %s <- o } }`, p.Input, p.Output),
		fmt.Sprintf("close(%s)", p.Output)
}

// Imports returns any extra import lines needed.
func (*Unslicer) Imports() []string {
	return nil
}

// RenameChannel renames any uses of the channel "from" to the channel "to".
func (p *Unslicer) RenameChannel(from, to string) {
	if p.Input == from {
		p.Input = to
	}
	if p.Output == from {
		p.Output = to
	}
}

// TypeKey returns the string "Unslicer"
func (*Unslicer) TypeKey() string {
	return "Unslicer"
}

// Update sets fields in the part based on info in the given Request.
func (p *Unslicer) Update(req *http.Request) error {
	if req == nil {
		return nil
	}
	if err := req.ParseForm(); err != nil {
		return err
	}
	p.Input = req.FormValue("Input")
	p.Output = req.FormValue("Output")
	return nil
}
