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
		c.Pins = make(map[NodePin]struct{})
	}
	for k, n := range g.Nodes {
		n.Name = k
		// Scan connections to cache them in channels.
		for p, co := range n.Connections {
			ch := g.Channels[co]
			if ch == nil {
				continue
			}
			ch.Pins[NodePin{Node: n.Name, Pin: p}] = struct{}{}
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
		g.Nodes[np.Node].Connections[np.Pin] = "nil"
	}
	delete(g.Channels, ch.Name)
}
