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
	"html/template"
	"net/http"
)

// TextFileReader waits for an input channel to close or send a value, then
// reads a file, and streams the lines of text to an output channel typed string,
// closing the output channel when done. If an error occurs, it stops reading and
// the error is sent to an error channel, which is not closed.
type TextFileReader struct {
	WaitFor string `json:"wait_for"`
	Path    string `json:"path"`
	Output  string `json:"output"`
	Error   string `json:"errors"`
}

// AssociateEditor associates a template called "part_view" with the given template.
func (r *TextFileReader) AssociateEditor(*template.Template) error {
	return nil
}

// Channels returns any channels used. Anything returned that is not a channel is ignored.
func (r *TextFileReader) Channels() (read, written []string) {
	return []string{r.WaitFor}, []string{r.Output}
}

// Clone returns a copy of this part.
func (r *TextFileReader) Clone() interface{} {
	s := *r
	return &s
}

// Impl returns Go source code implementing the part.
func (r *TextFileReader) Impl() (head, body, tail string) {
	if r.WaitFor != "" {
		head = fmt.Sprintf("<-%s", r.WaitFor)
	}
	body = fmt.Sprintf(`partlib.StreamTextFile("%s", %s, %s)`, r.Path, r.Output, r.Error)
	tail = fmt.Sprintf("close(%s)", r.Output)
	return head, body, tail
}

// Imports returns any extra import lines needed.
func (*TextFileReader) Imports() []string {
	return []string{
		`"github.com/google/shenzhen-go/parts/partlib"`,
	}
}

// Update sets fields in the part based on info in the given Request.
func (r *TextFileReader) Update(*http.Request) error {
	return nil
}

// TypeKey returns the string "TextFileReader"
func (*TextFileReader) TypeKey() string {
	return "TextFileReader"
}
