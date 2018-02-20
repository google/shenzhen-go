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

package graph

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/google/shenzhen-go/v0/parts"
	"github.com/google/shenzhen-go/v0/source"
)

// Part abstracts the implementation of a node. Concrete implementations should be
// able to be marshalled to and unmarshalled from JSON sensibly.
type Part interface {
	// AssociateEditor associates a template called "part_view" with the given template.
	AssociateEditor(*template.Template) error

	// Channels returns any additional channels the part thinks it uses.
	// This should be unnecessary.
	Channels() (read, written source.StringSet)

	// Clone returns a copy of this part.
	Clone() interface{}

	// Help returns a helpful description of what this part can do.
	Help() template.HTML

	// Impl returns Go source code implementing the part.
	// The head is executed, then the body is executed (# Multiplicity
	// instances of the body concurrently), then the tail (once the body/bodies
	// are finished).
	//
	// This allows cleanly closing channels for nodes with Multiplicity > 1.
	// The tail is deferred so that the body can use "return" and it is still
	// executed.
	Impl() (head, body, tail string)

	// Imports returns any extra import lines needed for the Part.
	Imports() []string

	// RenameChannel informs the part that a channel has been renamed.
	RenameChannel(from, to string)

	// TypeKey returns the "type" of part.
	TypeKey() string

	// Update sets fields in the part based on info in the given Request.
	Update(*http.Request) error
}

// PartFactory creates a part.
type PartFactory func() Part

// PartFactories translates part type strings into part factories.
var PartFactories = map[string]PartFactory{
	"Aggregator":     func() Part { return new(parts.Aggregator) },
	"Broadcast":      func() Part { return new(parts.Broadcast) },
	"Code":           func() Part { return new(parts.Code) },
	"Filter":         func() Part { return new(parts.Filter) },
	"HTTPServer":     func() Part { return new(parts.HTTPServer) },
	"StaticSend":     func() Part { return new(parts.StaticSend) },
	"TextFileReader": func() Part { return new(parts.TextFileReader) },
	"Unslicer":       func() Part { return new(parts.Unslicer) },
}

// Node models a goroutine. It can be marshalled and unmarshalled to JSON sensibly.
type Node struct {
	Part

	Name         string
	Multiplicity uint
	Wait         bool

	// Auto-extracted channel usage.
	chansRd, chansWr source.StringSet
}

func (n *Node) extractChans(defs string) error {
	h, b, t := n.Part.Impl()
	n.chansRd, n.chansWr = n.Part.Channels()
	hr, hw, err := source.ExtractChannels(h, "head", defs)
	if err != nil {
		return fmt.Errorf("extracting channels from head: %v", err)
	}
	br, bw, err := source.ExtractChannels(b, "body", defs)
	if err != nil {
		return fmt.Errorf("extracting channels from body: %v", err)
	}
	tr, tw, err := source.ExtractChannels(t, "tail", defs)
	if err != nil {
		return fmt.Errorf("extracting channels from tail: %v", err)
	}
	n.chansRd = source.Union(n.chansRd, hr, br, tr)
	n.chansWr = source.Union(n.chansWr, hw, bw, tw)
	return nil
}

// RenameChannel renames any uses of channel "from" to channel "to".
func (n *Node) RenameChannel(from, to string) {
	n.Part.RenameChannel(from, to)
	// Simple update of cached values
	if n.chansRd.Ni(from) {
		n.chansRd.Del(from)
		n.chansRd.Add(to)
	}
	if n.chansWr.Ni(from) {
		n.chansWr.Del(from)
		n.chansWr.Add(to)
	}
}

// Channels returns all the channels this node uses.
func (n *Node) Channels() (read, written source.StringSet) { return n.chansRd, n.chansWr }

// ChannelsRead returns the channels read from by this node. It is a convenience
// function for the templates, which can't do multiple returns.
func (n *Node) ChannelsRead() []string { return n.chansRd.Slice() }

// ChannelsWritten returns the channels written to by this node. It is a convenience
// function for the templates, which can't do multiple returns.
func (n *Node) ChannelsWritten() []string { return n.chansWr.Slice() }

// Copy returns a copy of this node, but with an empty name and a clone of the Part.
func (n *Node) Copy() *Node {
	return &Node{
		Name:         "",
		Multiplicity: n.Multiplicity,
		Wait:         n.Wait,
		Part:         n.Part.Clone().(Part),
	}
}

// ImplHead returns the Head part of the implementation.
func (n *Node) ImplHead() string {
	h, _, _ := n.Part.Impl()
	return h
}

// ImplBody returns the Body part of the implementation.
func (n *Node) ImplBody() string {
	_, b, _ := n.Part.Impl()
	return b
}

// ImplTail returns the Tail part of the implementation.
func (n *Node) ImplTail() string {
	_, _, t := n.Part.Impl()
	return t
}

func (n *Node) String() string { return n.Name }

type jsonNode struct {
	Name         string          `json:"name"`
	Wait         bool            `json:"wait"`
	Multiplicity uint            `json:"multiplicity"`
	Part         json.RawMessage `json:"part"`
	PartType     string          `json:"part_type"`
}

// MarshalJSON encodes the node and part as JSON.
func (n *Node) MarshalJSON() ([]byte, error) {
	p, err := json.Marshal(n.Part)
	if err != nil {
		return nil, err
	}
	if n.Multiplicity < 1 {
		n.Multiplicity = 1
	}
	return json.Marshal(&jsonNode{
		Part:         p,
		PartType:     n.Part.TypeKey(),
		Name:         n.Name,
		Wait:         n.Wait,
		Multiplicity: n.Multiplicity,
	})
}

// UnmarshalJSON decodes the node and part as JSON.
func (n *Node) UnmarshalJSON(j []byte) error {
	var mp jsonNode
	if err := json.Unmarshal(j, &mp); err != nil {
		return err
	}
	pf, ok := PartFactories[mp.PartType]
	if !ok {
		return fmt.Errorf("unknown part type %q", mp.PartType)
	}
	p := pf()
	if err := json.Unmarshal(mp.Part, p); err != nil {
		return err
	}
	if mp.Multiplicity < 1 {
		mp.Multiplicity = 1
	}
	n.Name = mp.Name
	n.Wait = mp.Wait
	n.Multiplicity = mp.Multiplicity
	n.Part = p
	return nil
}
