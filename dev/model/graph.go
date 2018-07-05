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
	// Finally, set up channel pin caches.
	g.RefreshChannelsPins()
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
// It doesn't fix conflicting names, but dedupes any whole lines.
// TODO: Put nodes in separate files to solve all import issues.
func (g *Graph) AllImports() []string {
	m := source.NewStringSet()
	m.Add(`"sync"`)
	for _, n := range g.Nodes {
		for _, i := range n.Part.Imports() {
			m.Add(i)
		}
	}
	return m.Slice()
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

// TypeIncompatibilityError is used when
type TypeIncompatibilityError struct {
	Summary string
	Source  error
}

func (e *TypeIncompatibilityError) Error() string {
	return e.Summary
}

// InferTypes resolves the types of channels and generic pins.
func (g *Graph) InferTypes() error {
	// All nodes start with no inferred type parameters, and all pin types
	// begin as their basic definition.
	for _, n := range g.Nodes {
		pins := n.Pins()
		n.typeParams = make(map[source.TypeParam]*source.Type)
		n.pinTypes = make(map[string]*source.Type, len(pins))
		for pn, p := range pins {
			pt, err := source.NewType(n.Name, p.Type)
			if err != nil {
				return err
			}
			n.pinTypes[pn] = pt
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
		//log.Printf("InferTypes: popping %s", c.Name)

		next, err := g.inferChannelType(c)
		if err != nil {
			return err
		}
		//log.Printf("InferTypes: pushing %d more", len(next))
		for c := range next {
			q = append(q, c)
		}
	}

	// Force all unresolved channel type parameters to interface{}.
	for _, c := range g.Channels {
		if c.Type == nil || c.Type.Plain() {
			continue
		}
		c.Type.Lithify(typeEmptyInterface)
	}
	// Force all unresolved node type parameters to interface{}.
	// Finally, give the node a map of resolved local types.
	for _, n := range g.Nodes {
		n.lithifyRemainingTypeParams()
		n.makeFinalTypes()
	}
	return nil
}

// next contains any channels that might be inferrable
// as a result of making improvement on this channel's type.
func (g *Graph) inferChannelType(c *Channel) (next map[*Channel]struct{}, err error) {
	next = make(map[*Channel]struct{})

	// Look at c's pins.
	for np := range c.Pins {
		n := g.Nodes[np.Node]
		tp := n.pinTypes[np.Pin]

		// Use it if we don't have a type already.
		if c.Type == nil {
			//log.Printf("inferChannelType: adopting %s immediately", tp)
			c.Type = tp
			next[c] = struct{}{}
			continue
		}

		// Cross-infer c.Type with tp.
		// c.Type refinement applies only to c.Type, but
		// tp refinement applies to all pins on node n!
		infc, err := c.Type.Infer(tp)
		if err != nil {
			return nil, &TypeIncompatibilityError{
				Summary: "channel connected to incompatible types",
				Source:  err,
			}
		}
		infp, err := tp.Infer(c.Type)
		if err != nil {
			return nil, &TypeIncompatibilityError{
				Summary: "channel connected to incompatible types",
				Source:  err,
			}
		}
		//log.Printf("inferChannelType: for pin %v inferred %v", np, inf)

		// Apply inferred params to c.Type.
		if _, err := c.Type.Refine(infc); err != nil {
			return nil, &TypeIncompatibilityError{
				Summary: "channel type refinement failed",
				Source:  err,
			}
		}

		// Apply inferred params to all pins on node n.
		nxcn, err := n.applyTypeParams(infp)
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
func (n *Node) applyTypeParams(types map[source.TypeParam]*source.Type) (next source.StringSet, err error) {
	// Merge the new map into n.typeParams, checking for mismatches.
	for par, typ := range types {
		old := n.typeParams[par]
		if old != nil {
			// So the old type exists. Is it compatible with the new type?
			if _, err := old.Infer(typ); err != nil {
				return nil, &TypeIncompatibilityError{
					Summary: "inferred type mismatch",
					Source:  err,
				}
			}
			if _, err := typ.Infer(old); err != nil {
				return nil, &TypeIncompatibilityError{
					Summary: "inferred type mismatch",
					Source:  err,
				}
			}
		}
		n.typeParams[par] = typ
	}
	// Refine all pin types.
	next = make(source.StringSet)
	for pn, pt := range n.pinTypes {
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

func (n *Node) lithifyRemainingTypeParams() {
	for _, pt := range n.pinTypes {
		for _, param := range pt.Params() {
			n.typeParams[param] = typeEmptyInterface
		}
		pt.Lithify(typeEmptyInterface)
	}
}

func (n *Node) makeFinalTypes() {
	n.finalTypeParams = make(map[string]string)
	for tp, typ := range n.typeParams {
		if tp.Scope != n.Name {
			continue
		}
		n.finalTypeParams[tp.Ident] = typ.String()
	}
}
