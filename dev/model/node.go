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
	"strings"
	"unicode"

	"github.com/google/shenzhen-go/dev/source"
)

// Node models a goroutine. This is the "real" model type for nodes.
// It can be marshalled and unmarshalled to JSON sensibly.
type Node struct {
	Part
	Name         string
	Comment      string
	Enabled      bool
	Multiplicity uint
	Wait         bool
	X, Y         float64
	Connections  map[string]string // Pin name -> channel name

	imports, head, body, tail string                  // Final implementation after type inference
	typeParams                map[string]string       // Local type parameter -> stringy type
	pinTypes                  map[string]*source.Type // Pin name -> inferred type of pin
}

// Copy returns a copy of this node, but with an empty name, nil connections, and a clone of the Part.
func (n *Node) Copy() *Node {
	n0 := &Node{
		Name:         "",
		Enabled:      n.Enabled,
		Multiplicity: n.Multiplicity,
		Wait:         n.Wait,
		Part:         n.Part.Clone().(Part),
		// TODO: find a better location
		X: n.X + 8,
		Y: n.Y + 100,
	}
	n0.RefreshConnections()
	return n0
}

// FlatImports returns the imports as a single string.
func (n *Node) FlatImports() string {
	return n.imports
}

// ImplHead returns the Head part of the implementation.
func (n *Node) ImplHead() string {
	return n.head
}

// ImplBody returns the Body part of the implementation.
func (n *Node) ImplBody() string {
	return n.body
}

// ImplTail returns the Tail part of the implementation.
func (n *Node) ImplTail() string {
	return n.tail
}

// PinFullTypes is a map from pin names to full resolved types:
// pinName <-chan someType or pinName chan<- someType.
// Requires InferTypes to have been called.
func (n *Node) PinFullTypes() map[string]string {
	pins := n.Pins()
	m := make(map[string]string, len(pins))
	for pn, p := range pins {
		m[pn] = p.Direction.Type() + " " + n.pinTypes[pn].String()
	}
	return m
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
	*PartJSON
	Comment      string            `json:"comment,omitempty"`
	Enabled      bool              `json:"enabled"`
	Wait         bool              `json:"wait"`
	Multiplicity uint              `json:"multiplicity,omitempty"`
	X            float64           `json:"x"`
	Y            float64           `json:"y"`
	Connections  map[string]string `json:"connections"`
}

// MarshalJSON encodes the node and part as JSON.
func (n *Node) MarshalJSON() ([]byte, error) {
	pj, err := MarshalPart(n.Part)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&jsonNode{
		PartJSON:     pj,
		Comment:      n.Comment,
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
	p, err := mp.PartJSON.Unmarshal()
	if err != nil {
		return err
	}
	if mp.Multiplicity < 1 {
		mp.Multiplicity = 1
	}
	n.Comment = mp.Comment
	n.Enabled = mp.Enabled
	n.Wait = mp.Wait
	n.Multiplicity = mp.Multiplicity
	n.Part = p
	n.X, n.Y = mp.X, mp.Y
	n.Connections = mp.Connections
	n.RefreshConnections()
	return nil
}

// RefreshConnections filters n.Connections to ensure only pins defined by the
// part are in the map, and any new ones are mapped to "nil".
func (n *Node) RefreshConnections() {
	pins := n.Pins()
	conns := make(map[string]string, len(pins))
	for name := range pins {
		c := n.Connections[name]
		if c == "" {
			c = "nil"
		}
		conns[name] = c
	}
	n.Connections = conns
}

// RefreshImpl refreshes imports, head, body, and tail, from the Part.
func (n *Node) RefreshImpl() {
	n.imports = strings.Join(n.Part.Imports(), "\n")
	n.head, n.body, n.tail = n.Part.Impl(n.typeParams)
}
