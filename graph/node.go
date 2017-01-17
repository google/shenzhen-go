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

	"github.com/google/shenzhen-go/parts"
)

// Part abstracts the implementation of a node. Concrete implementations should be
// able to be marshalled to and unmarshalled from JSON sensibly.
type Part interface {
	// AssociateEditor associates a template called "part_view" with the given template.
	AssociateEditor(*template.Template) error

	// Channels returns any channels used. Anything returned that is not a channel is ignored.
	Channels() (read, written []string)

	// Clone returns a copy of this part.
	Clone() interface{}

	// Impl returns Go source code implementing the part.
	// Loosely, the code will be inserted into the template as follows:
	//
	// 		go {
	// 		    {{head}}
	// 		    for n:=0; n<Multiplicity; n++ {
	//			    go {
	//					{{body}}
	//				}
	// 		    }
	// 		    {{tail}}
	// 		}
	//
	// This allows cleanly closing channels for nodes with Multiplicity > 1.
	Impl() (head, body, tail string)

	// Update sets fields in the part based on info in the given Request.
	Update(*http.Request) error

	// TypeKey returns the "type" of part.
	TypeKey() string
}

// PartFactory creates a part.
type PartFactory func() Part

// PartFactories translates part type strings into part factories.
var PartFactories = map[string]PartFactory{
	"Code":   func() Part { return new(parts.Code) },
	"Filter": func() Part { return new(parts.Filter) },
	// "Multiplexer": func() Part { return new(parts.Multiplexer) },
}

// Node models a goroutine. It can be marshalled and unmarshalled to JSON sensibly.
type Node struct {
	Part

	Name         string
	Multiplicity uint
	Wait         bool
}

// ChannelsRead returns the channels read from by this node. It is a convenience
// function for the templates, which can't do multiple returns.
func (n *Node) ChannelsRead() []string {
	r, _ := n.Part.Channels()
	return r
}

// ChannelsWritten returns the channels written to by this node. It is a convenience
// function for the templates, which can't do multiple returns.
func (n *Node) ChannelsWritten() []string {
	_, w := n.Part.Channels()
	return w
}

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
	return p.Update(nil)
}
