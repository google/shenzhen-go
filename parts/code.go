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
	"log"
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
	// TODO: go back to storing strings, implement marshal/unmarshal JSON.

	Head []string `json:"head"`
	Body []string `json:"body"`
	Tail []string `json:"tail"`

	// Computed from Head + Body + Tail - which channels are read from and written to.
	chansRd, chansWr source.StringSet
}

// AssociateEditor adds a "part_view" template to the given template.
func (c *Code) AssociateEditor(tmpl *template.Template) error {
	_, err := tmpl.New("part_view").Parse(codePartEditTemplateSrc)
	return err
}

// Channels returns the names of all channels used by this goroutine.
func (c *Code) Channels() (read, written source.StringSet) { return c.chansRd, c.chansWr }

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

// RenameChannel does fancy footwork to rename the channel in the code,
// with a side-effect of nicely formatting it. If a rename issue occurs
// e.g. because the user's code has a syntax error, the rename is aborted
// and logged.
func (c *Code) RenameChannel(from, to string) {
	h, b, t := c.Impl()
	h1, err := source.RenameIdent(h, "head", from, to)
	if err != nil {
		log.Printf("Couldn't do rename on head: %v", err)
		return
	}
	b1, err := source.RenameIdent(b, "body", from, to)
	if err != nil {
		log.Printf("Couldn't do rename on body: %v", err)
		return
	}
	t1, err := source.RenameIdent(t, "tail", from, to)
	if err != nil {
		log.Printf("Couldn't do rename on tail: %v", err)
		return
	}
	c.Head = strings.Split(h1, "\n")
	c.Body = strings.Split(b1, "\n")
	c.Tail = strings.Split(t1, "\n")

	// Simple update of cached values
	if c.chansRd.Ni(from) {
		c.chansRd.Del(from)
		c.chansRd.Add(to)
	}
	if c.chansWr.Ni(from) {
		c.chansWr.Del(from)
		c.chansWr.Add(to)
	}
}

// TypeKey returns "Code".
func (*Code) TypeKey() string { return "Code" }

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
	c.chansRd = source.Union(hs, bs, ts)
	c.chansWr = source.Union(hd, bd, td)
	return nil
}

func stripCR(in []string) {
	for i := range in {
		in[i] = strings.TrimSuffix(in[i], "\r")
	}
}
