// Copyright 2016 Google Inc.
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
	"github.com/google/shenzhen-go/source"
	"html/template"
	"net/http"
)

const textFileReaderEditTemplateSrc = `
<div class="formfield">
	<label for="PathInput">File paths to read</label>
	<select name="PathInput">
		{{range .Graph.Channels -}}
		{{if eq .Type "string" }}
		<option value="{{.Name}}" {{if eq .Name $.Node.Part.PathInput}}selected{{end}}>{{.Name}}</option>
		{{- end}}
		{{- end}}
	</select>
</div>
<div class="formfield">
	<label for="Output">Output</label>
	<select name="Output">
		{{range .Graph.Channels -}}
		{{if eq .Type "partlib.FileLine" }}
		<option value="{{.Name}}" {{if eq .Name $.Node.Part.Output}}selected{{end}}>{{.Name}}</option>
		{{- end}}
		{{- end}}
	</select>
</div>
<div class="formfield">
	<label for="Error">Error</label>
	<select name="Error">
		{{range .Graph.Channels -}}
		{{if eq .Type "error"}}
		<option value="{{.Name}}" {{if eq .Name $.Node.Part.Error}}selected{{end}}>{{.Name}}</option>
		{{- end}}
		{{- end}}
	</select>
</div>
`

// TextFileReader waits for the path of a file to read to arrive, the
// reads the file, and streams the lines of text to an output channel typed string,
// closing the output channel when done. If an error occurs, it stops reading and
// the error is sent to an error channel, which is not closed.
type TextFileReader struct {
	PathInput string `json:"path_input"`
	Output    string `json:"output"`
	Error     string `json:"errors"`
}

// AssociateEditor associates a template called "part_view" with the given template.
func (r *TextFileReader) AssociateEditor(t *template.Template) error {
	_, err := t.New("part_view").Parse(textFileReaderEditTemplateSrc)
	return err
}

// Channels returns any channels used. Anything returned that is not a channel is ignored.
func (r *TextFileReader) Channels() (read, written source.StringSet) {
	return source.NewStringSet(r.PathInput), source.NewStringSet(r.Output, r.Error)
}

// Clone returns a copy of this part.
func (r *TextFileReader) Clone() interface{} {
	s := *r
	return &s
}

// Help returns useful help information.
func (*TextFileReader) Help() template.HTML {
	return `<p>
	The TextFileReader waits until it receives a path to a file (File paths to read).
	For each path, it tries to open the file at that path, separate it into multiple lines of text,
	and send each line of text to the output (as a value of type <code>partlib.FileLine</code>).
	When the channel with file-paths is closed, the output channel will be closed (once all the files have been read).
	</p><p>
	If an error occurs, the error is sent to the error channel. An error can occur if the file couldn't be opened,
	the file couldn't be split into lines of text, or the file couldn't be closed.
	</p><p>
	The Multiplicity parameter can be used to allow the TextFileReader to read multiple files concurrently.
	</p>`
}

// Impl returns Go source code implementing the part.
func (r *TextFileReader) Impl() (head, body, tail string) {
	body = fmt.Sprintf(`partlib.StreamTextFile(%s, %s, %s)`, r.PathInput, r.Output, r.Error)
	tail = fmt.Sprintf("close(%s)", r.Output)
	return "", body, tail
}

// Imports returns any extra import lines needed.
func (*TextFileReader) Imports() []string {
	return []string{
		`"github.com/google/shenzhen-go/parts/partlib"`,
	}
}

// RenameChannel renames any uses of the channel "from" to the channel "to".
func (r *TextFileReader) RenameChannel(from, to string) {
	if r.PathInput == from {
		r.PathInput = to
	}
	if r.Output == from {
		r.Output = to
	}
	if r.Error == from {
		r.Output = to
	}
}

// TypeKey returns the string "TextFileReader"
func (*TextFileReader) TypeKey() string { return "TextFileReader" }

// Update sets fields in the part based on info in the given Request.
func (r *TextFileReader) Update(req *http.Request) error {
	if req == nil {
		return nil
	}
	if err := req.ParseForm(); err != nil {
		return err
	}
	r.PathInput = req.FormValue("PathInput")
	r.Output = req.FormValue("Output")
	r.Error = req.FormValue("Error")
	return nil
}
