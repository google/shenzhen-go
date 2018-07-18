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

package model

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/google/shenzhen-go/dev/source"
)

var typeEmptyInterface = source.MustNewType("", "interface{}")

// Graph represents a package / program / collection of nodes and channels.
type Graph struct {
	FilePath    string              `json:"-"` // path to the JSON source
	URLPath     string              `json:"-"` // path in the URL
	Name        string              `json:"name"`
	PackagePath string              `json:"package_path"`
	IsCommand   bool                `json:"is_command"`
	Nodes       map[string]*Node    `json:"nodes"`    // name -> node
	Channels    map[string]*Channel `json:"channels"` // name -> channel

	types source.TypeInferenceMap
}

// NewGraph returns a new empty graph associated with a file path.
func NewGraph(filePath, urlPath, pkgPath string) *Graph {
	return &Graph{
		FilePath:    filePath,
		URLPath:     urlPath,
		PackagePath: pkgPath,
		Channels:    make(map[string]*Channel),
		Nodes:       make(map[string]*Node),
	}
}

// LoadJSON loads a JSON-encoded Graph from an io.Reader.
func LoadJSON(r io.Reader, filePath, urlPath string) (*Graph, error) {
	dec := json.NewDecoder(r)
	g := &Graph{
		FilePath: filePath,
		URLPath:  urlPath,
	}
	if err := dec.Decode(g); err != nil {
		return nil, err
	}
	// Each node and channel should cache it's own name.
	for k, c := range g.Channels {
		c.Name = k
	}
	for k, n := range g.Nodes {
		n.Name = k
	}
	// Set up channel pin caches.
	g.RefreshChannelsPins()
	// As a safety mechanism, cull any connections to channels that don't exist.
	for _, n := range g.Nodes {
		for p, co := range n.Connections {
			if g.Channels[co] == nil {
				n.Connections[p] = "nil"
			}
		}
	}
	return g, nil
}

// PackageName extracts the name of the package from the package path ("full" package name).
func (g *Graph) PackageName() string {
	i := strings.LastIndex(g.PackagePath, "/")
	if i < 0 {
		return g.PackagePath
	}
	return g.PackagePath[i+1:]
}

// AllImports combines all desired imports into one slice.
// It doesn't fix conflicting names, but dedupes any whole lines,
// trims whitespace and removes blank lines. go/format will put
// them in sorted order later.
// TODO: Put nodes in separate files to solve all import issues.
func (g *Graph) AllImports() []string {
	m := source.NewStringSet(`"runtime"`, `"sync"`)
	for _, n := range g.Nodes {
		for _, i := range n.Impl.Imports {
			j := strings.TrimSpace(i)
			if j == "" {
				continue
			}
			m.Add(j)
		}
	}
	return m.Slice()
}

// Inits returns a map of part type keys to init sections for those parts that need it.
func (g *Graph) Inits() map[string]string {
	m := make(map[string]string)
	for _, n := range g.Nodes {
		if !n.Impl.NeedsInit {
			continue
		}
		k := n.Part.TypeKey()
		i := PartTypes[k].Init
		if i == "" {
			continue
		}
		m[k] = i
	}
	return m
}

// DeleteChannel cleans up any connections and then deletes a channel.
func (g *Graph) DeleteChannel(ch *Channel) {
	for np := range ch.Pins {
		n := g.Nodes[np.Node]
		if n == nil {
			panic("node " + np.Node + " should exist")
		}
		n.Connections[np.Pin] = "nil"
	}
	delete(g.Channels, ch.Name)
}

// DeleteNode cleans up any connections and then deletes a node.
// If cleanupChans is set, it then deletes any channels which have less
// than 2 pins left as a result of deleting the node.
func (g *Graph) DeleteNode(n *Node, cleanupChans bool) {
	// In case the node is connected to the same channel more than
	// once, first disconnect all the pins, then delete any channels.
	var rem []*Channel
	for p, cn := range n.Connections {
		if cn == "nil" {
			continue
		}
		ch := g.Channels[cn]
		if ch == nil {
			continue
		}
		ch.RemovePin(n.Name, p)
		if cleanupChans && len(ch.Pins) < 2 {
			rem = append(rem, ch)
		}
	}
	delete(g.Nodes, n.Name)
	for _, ch := range rem {
		g.DeleteChannel(ch)
	}
}

// RenameNode renames the node and fixes up references.
func (g *Graph) RenameNode(n *Node, newName string) {
	if newName == n.Name {
		return
	}
	// Fix any channels - not too fiddly.
	for p, co := range n.Connections {
		if co == "nil" {
			continue
		}
		ch := g.Channels[co]
		if ch == nil {
			continue
		}
		ch.RemovePin(n.Name, p)
		ch.AddPin(newName, p)
	}
	// Update the nodes map.
	delete(g.Nodes, n.Name)
	g.Nodes[newName] = n
	n.Name = newName
}

// Check checks over the graph for any errors.
func (g *Graph) Check() error {
	// TODO: implement
	return errors.New("not implemented")
}

// RefreshChannelsPins refreshes the Pins cache of all channels.
// Use this when node names or pin definitions might have changed.
func (g *Graph) RefreshChannelsPins() {
	// Reset all caches.
	for _, ch := range g.Channels {
		ch.Pins = make(map[NodePin]struct{})
	}
	// Add only those that now exist.
	for _, n := range g.Nodes {
		for p, co := range n.Connections {
			ch := g.Channels[co]
			if ch == nil {
				continue
			}
			ch.AddPin(n.Name, p)
		}
	}
	// Check for channels with < 2 pins.
	for _, ch := range g.Channels {
		if len(ch.Pins) < 2 {
			g.DeleteChannel(ch)
		}
	}
}

// TypeIncompatibilityError is used when types mismatch during inference.
// TODO(josh): Make this useful to feed into the error panel.
type TypeIncompatibilityError struct {
	Summary string
	Source  error
}

func (e *TypeIncompatibilityError) Error() string {
	return e.Summary
}

// InferTypes resolves the types of channels and generic pins.
func (g *Graph) InferTypes() error {
	// The graph starts with no inferred types, and all pin types
	// begin as their basic definition, params scoped to the node.
	// The types map should start with all type parameters set to nil.
	g.types = make(source.TypeInferenceMap)
	for _, n := range g.Nodes {
		pins := n.Part.Pins()
		n.PinTypes = make(map[string]*source.Type, len(pins))
		for pn, p := range pins {
			pt, err := source.NewType(n.Name, p.Type)
			if err != nil {
				return err
			}
			n.PinTypes[pn] = pt
			g.types.Note(pt)
		}
	}

	// Construct a queue of channels to resolve, and reset channel types.
	q := make([]*Channel, 0, len(g.Channels))
	for _, c := range g.Channels {
		c.Type = nil
		q = append(q, c)
	}

	// Flood fill inference.
	for len(q) > 0 {
		c := q[0]
		q = q[1:]

		next, err := g.inferAndRefineChan(c)
		if err != nil {
			return err
		}
		for c := range next {
			q = append(q, c)
		}
	}

	g.types.ApplyDefault(typeEmptyInterface)

	// Refine all types one final time.
	for _, c := range g.Channels {
		c.Type.Refine(g.types)
	}
	for _, n := range g.Nodes {
		// Some pins aren't connected, lithify those too.
		for _, pt := range n.PinTypes {
			pt.Refine(g.types)
		}
		n.TypeParams = make(map[string]*source.Type)
	}
	// Finally, give each node a limited view of relevant inferred types.
	for tp, typ := range g.types {
		g.Nodes[tp.Scope].TypeParams[tp.Ident] = typ
	}
	return nil
}

// next contains any channels that might be inferrable
// as a result of making improvement on this channel's type.
func (g *Graph) inferAndRefineChan(c *Channel) (map[*Channel]struct{}, error) {
	next := make(map[*Channel]struct{})

	// Look at c's pins.
	for np := range c.Pins {
		n := g.Nodes[np.Node]
		ptype := n.PinTypes[np.Pin]

		// Use ptype for c.Type if nothing else.
		if c.Type == nil {
			c.Type = ptype
			next[c] = struct{}{}
			continue
		}

		// Make inferences; at the end, c.Type and ptype must be the same fully refined type.
		if err := g.types.Infer(c.Type, ptype); err != nil {
			return nil, &TypeIncompatibilityError{
				Summary: "channel connected to incompatible types",
				Source:  err,
			}
		}

		// Apply inferred params back to c.Type.
		cimp, err := c.Type.Refine(g.types)
		if err != nil {
			return nil, &TypeIncompatibilityError{
				Summary: "channel type refinement failed",
				Source:  err,
			}
		}
		if cimp {
			next[c] = struct{}{}
		}

		// Apply inferred params to all pins on node n.
		nxcn, err := n.applyTypeParams(g.types)
		if err != nil {
			return nil, err
		}
		// Push potentially-affected channels.
		for cn := range nxcn {
			next[g.Channels[cn]] = struct{}{}
		}

	}
	return next, nil
}

// next is the names of channels that might be inferrable as a result of this apply.
func (n *Node) applyTypeParams(types source.TypeInferenceMap) (next source.StringSet, err error) {
	// Refine all pin types.
	next = make(source.StringSet)
	for pn, pt := range n.PinTypes {
		changed, err := pt.Refine(types)
		if err != nil {
			return nil, err
		}
		if !changed { // Refine had no effect, not worth investigating channel.
			continue
		}
		ch := n.Connections[pn]
		if ch == "" || ch == "nil" {
			continue
		}
		next.Add(ch)
	}
	return next, nil
}
