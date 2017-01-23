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
	"strings"

	"github.com/google/shenzhen-go/source"
)

const codePartEditTemplateSrc = `
<h4>Head</h4>
<textarea name="Head" rows="{{len .Node.Part.Head}}" cols="80">{{.Node.ImplHead}}</textarea>
<h4>Body</h4>
<textarea name="Body" rows="{{len .Node.Part.Body}}" cols="80">{{.Node.ImplBody}}</textarea>
<h4>Tail</h4>
<textarea name="Tail" rows="{{len .Node.Part.Tail}}" cols="80">{{.Node.ImplTail}}</textarea>
`

// Code is a component containing arbitrary code.
type Code struct {
	// You may be wondering why {Head, Body, Tail} aren't just typed "string"
	// instead of []string.
	// It used to be, but then the JSON file would include blobs of strings with
	// no separation of lines.
	// encoding/json still escapes things like \t and \u003c, but whatevs.

	Head []string `json:"head"`
	Body []string `json:"body"`
	Tail []string `json:"tail"`

	// Computed from Head + Body + Tail - which channels are read from and written to.
	chansRd, chansWr []string
}

// AssociateEditor adds a "part_view" template to the given template.
func (c *Code) AssociateEditor(tmpl *template.Template) error {
	_, err := tmpl.New("part_view").Parse(codePartEditTemplateSrc)
	return err
}

// Channels returns the names of all channels used by this goroutine.
func (c *Code) Channels() (read, written []string) { return c.chansRd, c.chansWr }

// Clone returns a copy of this Code part.
func (c *Code) Clone() interface{} {
	return &Code{
		Head:    append([]string(nil), c.Head...),
		Body:    append([]string(nil), c.Body...),
		Tail:    append([]string(nil), c.Tail...),
		chansRd: c.chansRd,
		chansWr: c.chansWr,
	}
}

// Impl returns the implementation of the goroutine.
func (c *Code) Impl() (head, body, tail string) {
	return strings.Join(c.Head, "\n"),
		strings.Join(c.Body, "\n"),
		strings.Join(c.Tail, "\n")
}

// Imports returns a nil slice.
func (*Code) Imports() []string { return nil }

// Update sets relevant fields based on the given Request.
func (c *Code) Update(r *http.Request) error {
	// TODO: Do this less long-windedly
	h, b, t := c.Impl()
	if r != nil {
		h, b, t = r.FormValue("Head"), r.FormValue("Body"), r.FormValue("Tail")
	}
	hs, hd, err := source.ExtractChannelIdents(h, "head")
	if err != nil {
		return fmt.Errorf("extracting channels from head: %v", err)
	}
	bs, bd, err := source.ExtractChannelIdents(b, "body")
	if err != nil {
		return fmt.Errorf("extracting channels from body: %v", err)
	}
	ts, td, err := source.ExtractChannelIdents(t, "tail")
	if err != nil {
		return fmt.Errorf("extracting channels from tail: %v", err)
	}
	c.Head = strings.Split(h, "\n")
	stripCR(c.Head)
	c.Body = strings.Split(b, "\n")
	stripCR(c.Body)
	c.Tail = strings.Split(t, "\n")
	stripCR(c.Tail)
	c.chansRd = append(append(hs, bs...), ts...)
	c.chansWr = append(append(hd, bd...), td...)
	return nil
}

// TypeKey returns "Code".
func (*Code) TypeKey() string { return "Code" }

func stripCR(in []string) {
	for i := range in {
		in[i] = strings.TrimSuffix(in[i], "\r")
	}
}
