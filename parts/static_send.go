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
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/google/shenzhen-go/source"
)

const (
	staticSendEditorTemplateSrc = `
<div class="formfield">
	<label for="Output">Output</label>
	<select name="Output">
		{{range .Graph.Channels -}}
		<option value="{{.Name}}" {{if eq .Name $.Node.Part.Output}}selected{{end}}>{{.Name}}</option>
		{{- end}}
	</select>
</div>
<div class="formfield">
	<label for="Items">Items</label>
	<textarea name="Items" rows="{{len $.Node.Part.Items}}" cols="80">{{$.Node.Part.AllItems}}</textarea>
</div>
`
)

// StaticSend is a mostly blank template for a part, to make it easier to write more parts.
// TODO: Add necessary fields and write a good doc comment.
type StaticSend struct {
	Output string   `json:"output"`
	Items  []string `json:"items"`
}

// AllItems returns all the Items in a single string.
func (p *StaticSend) AllItems() string {
	return strings.Join(p.Items, "\n")
}

// AssociateEditor associates a template called "part_view" with the given template.
func (*StaticSend) AssociateEditor(t *template.Template) error {
	_, err := t.New("part_view").Parse(staticSendEditorTemplateSrc)
	return err
}

// Channels returns any channels used. Anything returned that is not a channel is ignored.
func (p *StaticSend) Channels() (read, written source.StringSet) {
	return nil, source.NewStringSet(p.Output)
}

// Clone returns a copy of this part.
func (p *StaticSend) Clone() interface{} {
	// TODO: Make sure Clone does the right thing.
	s := *p
	return &s
}

// Help returns useful help information.
func (*StaticSend) Help() template.HTML {
	return `<p>
	StaticSend writes a sequence of values to a channel, then closes it.
	</p><p>
	Enter one value per line. A value can be any Go literal or expression.
	</p><p>
	For example, the items:
	<p>
	<code>2<br>3<br>5<br>"melon"<br></code>
	</p>
	will send the numbers 2, 3, 5, and the string "melon" onto the output channel, in that order.
	</p>`
}

// Impl returns Go source code implementing the part.
func (p *StaticSend) Impl() (head, body, tail string) {
	buf := new(bytes.Buffer)
	for _, i := range p.Items {
		i = strings.TrimSpace(i)
		// Preserve comments and blank lines.
		if i == "" || strings.HasPrefix(i, "//") {
			fmt.Fprintln(buf, i)
			continue
		}
		fmt.Fprintf(buf, "%s <- %s\n", p.Output, i)
	}
	return "", buf.String(), fmt.Sprintf("close(%s)", p.Output)
}

// Imports returns any extra import lines needed.
func (*StaticSend) Imports() []string {
	return nil
}

// RenameChannel renames any uses of the channel "from" to the channel "to".
func (p *StaticSend) RenameChannel(from, to string) {
	if p.Output == from {
		p.Output = to
	}
}

// TypeKey returns the string "StaticSend"
func (*StaticSend) TypeKey() string {
	return "StaticSend"
}

// Update sets fields in the part based on info in the given Request.
func (p *StaticSend) Update(req *http.Request) error {
	if req == nil {
		return nil
	}
	if err := req.ParseForm(); err != nil {
		return err
	}
	p.Output = req.FormValue("Output")
	p.Items = strings.Split(req.FormValue("Items"), "\n")
	stripCR(p.Items)
	return nil
}
