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
// This will not result in deleting channels that are no longer valid.
func (g *Graph) DeleteNode(n *Node) {
	for p, cn := range n.Connections {
		if cn == "nil" {
			continue
		}
		ch := g.Channels[cn]
		if ch == nil {
			panic("channel " + cn + " should exist")
		}
		ch.RemovePin(n.Name, p)
	}
	delete(g.Nodes, n.Name)
}

// Check checks over the graph for any errors.
func (g *Graph) Check() error {
	// TODO: implement
	return errors.New("not implemented")
}

// RefreshChannelsPins refreshes the Pins cache of all channels.
// Use this when pin definitions might have changed.
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
