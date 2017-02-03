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
	"encoding/json"
	"fmt"
	"go/format"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/google/shenzhen-go/source"
)

const codePartEditTemplateSrc = `{{$lines := .Node.Part.LineCount}}
<h4>Head</h4>
<textarea name="Head" rows="{{$lines.H}}" cols="80">{{.Node.ImplHead}}</textarea>
<h4>Body</h4>
<textarea name="Body" rows="{{$lines.B}}" cols="80">{{.Node.ImplBody}}</textarea>
<h4>Tail</h4>
<textarea name="Tail" rows="{{$lines.T}}" cols="80">{{.Node.ImplTail}}</textarea>
`

// Code is a component containing arbitrary code.
type Code struct {
	head, body, tail string

	// Computed from Head + Body + Tail - which channels are read from and written to.
	chansRd, chansWr source.StringSet
}

type jsonCode struct {
	Head []string `json:"head"`
	Body []string `json:"body"`
	Tail []string `json:"tail"`
}

// MarshalJSON encodes the Code component as JSON.
func (c *Code) MarshalJSON() ([]byte, error) {
	k := &jsonCode{
		Head: strings.Split(c.head, "\n"),
		Body: strings.Split(c.body, "\n"),
		Tail: strings.Split(c.tail, "\n"),
	}
	stripCR(k.Head)
	stripCR(k.Body)
	stripCR(k.Tail)
	return json.Marshal(k)
}

// UnmarshalJSON decodes the Code component from JSON.
func (c *Code) UnmarshalJSON(j []byte) error {
	var mp jsonCode
	if err := json.Unmarshal(j, &mp); err != nil {
		return err
	}
	h := strings.Join(mp.Head, "\n")
	b := strings.Join(mp.Body, "\n")
	t := strings.Join(mp.Tail, "\n")
	if err := c.refresh(h, b, t); err != nil {
		// refresh would error if it can't get used channels,
		// say, because of a syntax error preventing parsing.
		// Returning error would stop the graph being loaded,
		// and the user might need to fix their syntax error,
		// via the interface...
		log.Printf("Couldn't format or determine channels used: %v", err)
	}
	return nil
}

// LineCount is number of lines in c.head, c.body, c.tail
// (conveneince function for templates.)
func (c *Code) LineCount() struct{ H, B, T int } {
	return struct{ H, B, T int }{
		H: strings.Count(c.head, "\n") + 1,
		B: strings.Count(c.body, "\n") + 1,
		T: strings.Count(c.tail, "\n") + 1,
	}
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
		head:    c.head,
		body:    c.body,
		tail:    c.tail,
		chansRd: source.Union(c.chansRd),
		chansWr: source.Union(c.chansWr),
	}
}

// Help returns a helpful explanation.
func (*Code) Help() template.HTML {
	return `<p>
	A Code part runs (executes) any Go code that you write.
	</p><p>
	It consists of 3 parts: a head, a body, and a tail. 
	The head runs first, and only runs once, no matter what number Multiplicity is set to. 
	The body runs next. The number of concurrent copies of the body that run is set by Multiplicity.
	Finally, when all copies of the body return, the tail runs. 
	</p><p>
	The head and tail are useful for operations that should only be done once. For example, any 
	output channels written to in the body can be correctly closed (if desired) in the tail.
	</p><p>
	Each instance of the body can use the int parameters <code>instanceNumber</code> and <code>multiplicity</code>
	to distinguish which instance is running and how many are running, if necessary. 
	<code>0 <= instanceNumber < multiplicity</code>
	</p><p>
	Any channels referred to will automatically be detected and shown in the graph, and
	when channels are renamed, these will be safely updated in the Code where they are
	referred to.
	</p><p>
	The <code>return</code> statement is allowed but optional in Code. There are no values that
	need to be returned.
	Using <code>return</code> in the head will prevent the body or tail from executing, but 
	using <code>return</code> in the body won't affect whether the tail is executed.
	</p>
	`
}

// Impl returns the implementation of the goroutine.
func (c *Code) Impl() (head, body, tail string) {
	return c.head, c.body, c.tail
}

// Imports returns a nil slice.
func (*Code) Imports() []string { return nil }

// RenameChannel does fancy footwork to rename the channel in the code,
// with a side-effect of nicely formatting it. If a rename issue occurs
// e.g. because the user's code has a syntax error, the rename is aborted
// and logged.
func (c *Code) RenameChannel(from, to string) {
	h, err := source.RenameChannel(c.head, "head", from, to)
	if err != nil {
		log.Printf("Couldn't do rename on head: %v", err)
		return
	}
	b, err := source.RenameChannel(c.body, "body", from, to)
	if err != nil {
		log.Printf("Couldn't do rename on body: %v", err)
		return
	}
	t, err := source.RenameChannel(c.tail, "tail", from, to)
	if err != nil {
		log.Printf("Couldn't do rename on tail: %v", err)
		return
	}
	c.head, c.body, c.tail = h, b, t

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

func (c *Code) refresh(h, b, t string) error {
	// At least save what the user entered.
	c.head, c.body, c.tail = h, b, t

	// It can probably have channels extracted...
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
	c.chansRd = source.Union(hs, bs, ts)
	c.chansWr = source.Union(hd, bd, td)

	// Try to format it.
	hf, err := format.Source([]byte(h))
	if err != nil {
		return err
	}
	bf, err := format.Source([]byte(b))
	if err != nil {
		return err
	}
	tf, err := format.Source([]byte(t))
	if err != nil {
		return err
	}

	c.head, c.body, c.tail = string(hf), string(bf), string(tf)
	return nil
}

// Update sets relevant fields based on the given Request.
func (c *Code) Update(r *http.Request) error {
	h, b, t := c.head, c.body, c.tail
	if r != nil {
		h, b, t = r.FormValue("Head"), r.FormValue("Body"), r.FormValue("Tail")
	}
	return c.refresh(h, b, t)
}
