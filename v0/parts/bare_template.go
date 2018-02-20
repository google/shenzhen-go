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
	"html/template"
	"net/http"

	"github.com/google/shenzhen-go/v0/source"
)

// This is a template for constructing new types of parts.
// It's mostly empty - lots of bits to fill in here and there.

// TODO: Define a useful editor template in here.
const partEditorTemplateSrc = `
<div class="formfield">
	<label for="Something">Something</label>
	<select name="Something">
		{{range .Graph.Channels -}}
		{{if eq .Type "interface{}" }}
		<option value="{{.Name}}" {{if eq .Name $.Node.Part.Something}}selected{{end}}>{{.Name}}</option>
		{{- end}}
		{{- end}}
	</select>
</div>
`

// PartTemplate is a mostly blank template for a part, to make it easier to write more parts.
// TODO: Add necessary fields and write a good doc comment.
type PartTemplate struct {
	Something string `json:"something"`
}

// AssociateEditor associates a template called "part_view" with the given template.
func (*PartTemplate) AssociateEditor(t *template.Template) error {
	// TODO: Make sure the template is correct here.
	_, err := t.New("part_view").Parse(partEditorTemplateSrc)
	return err
}

// Channels returns any channels used. Anything returned that is not a channel is ignored.
func (p *PartTemplate) Channels() (read, written source.StringSet) {
	return nil, nil // TODO: Return any channels used here.
}

// Clone returns a copy of this part.
func (p *PartTemplate) Clone() interface{} {
	// TODO: Make sure Clone does the right thing.
	s := *p
	return &s
}

// Help returns useful help information.
func (*PartTemplate) Help() template.HTML {
	return `<blink><h1>TODO</h1></blink>` // TODO: Return helpful information here
}

// Impl returns Go source code implementing the part.
func (p *PartTemplate) Impl() (head, body, tail string) {
	// TODO: Return some Go source code here
	return "", "", ""
}

// Imports returns any extra import lines needed.
func (*PartTemplate) Imports() []string {
	return nil // TODO: Return any necessary imports here
}

// RenameChannel renames any uses of the channel "from" to the channel "to".
func (p *PartTemplate) RenameChannel(from, to string) {
	// TODO: Safely update channels that are used here
}

// TypeKey returns the string "PartTemplate"
func (*PartTemplate) TypeKey() string {
	return "PartTemplate" // TODO: Return the right string here
}

// Update sets fields in the part based on info in the given Request.
func (p *PartTemplate) Update(req *http.Request) error {
	if req == nil {
		return nil
	}
	if err := req.ParseForm(); err != nil {
		return err
	}
	// TODO: Implement your part update here
	return nil
}
