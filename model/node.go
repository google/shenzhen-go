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

package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/shenzhen-go/model/parts"
)

// PartFactory creates a part.
type PartFactory func() Part

// PartFactories translates part type strings into part factories.
var PartFactories = map[string]PartFactory{
	/*	"Aggregator":     func() Part { return new(parts.Aggregator) },
		"Broadcast":      func() Part { return new(parts.Broadcast) }, */
	"Code": func() Part { return new(parts.Code) },
	/*	"Filter":         func() Part { return new(parts.Filter) },
		"HTTPServer":     func() Part { return new(parts.HTTPServer) },
		"StaticSend":     func() Part { return new(parts.StaticSend) },
		"TextFileReader": func() Part { return new(parts.TextFileReader) },
		"Unslicer":       func() Part { return new(parts.Unslicer) },*/
}

// Node models a goroutine. This is the "real" model type for nodes.
// It can be marshalled and unmarshalled to JSON sensibly.
type Node struct {
	Part
	Name         string
	Enabled      bool
	Multiplicity uint
	Wait         bool
	X, Y         int
	Connections  map[string]string // Pin name -> channel name
}

// Copy returns a copy of this node, but with an empty name and a clone of the Part.
func (n *Node) Copy() *Node {
	return &Node{
		Name:         "",
		Enabled:      n.Enabled,
		Multiplicity: n.Multiplicity,
		Wait:         n.Wait,
		Part:         n.Part.Clone().(Part),
		// TODO: find a better location
		X: n.X + 8,
		Y: n.Y + 100,
	}
}

// FlatImports returns the imports as a single string.
func (n *Node) FlatImports() string {
	return strings.Join(n.Part.Imports(), "\n")
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

// Identifier turns the name into a similar-looking identifier.
func (n *Node) Identifier() string {
	base := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '_'
		}
		if !unicode.IsLetter(r) && r != '_' && !unicode.IsDigit(r) {
			return -1
		}
		return r
	}, n.Name)
	var f rune
	for _, r := range base {
		f = r
		break
	}
	if unicode.IsDigit(f) {
		base = "_" + base
	}
	return base
}

type jsonNode struct {
	Enabled      bool              `json:"enabled"`
	Wait         bool              `json:"wait"`
	Multiplicity uint              `json:"multiplicity"`
	Part         json.RawMessage   `json:"part,omitempty"`
	PartType     string            `json:"part_type,omitempty"`
	X            int               `json:"x"`
	Y            int               `json:"y"`
	Connections  map[string]string `json:"connections"`
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
		Enabled:      n.Enabled,
		Wait:         n.Wait,
		Multiplicity: n.Multiplicity,
		X:            n.X,
		Y:            n.Y,
		Connections:  n.Connections,
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
	n.Enabled = mp.Enabled
	n.Wait = mp.Wait
	n.Multiplicity = mp.Multiplicity
	n.Part = p
	n.X, n.Y = mp.X, mp.Y
	n.Connections = mp.Connections
	return nil
}
