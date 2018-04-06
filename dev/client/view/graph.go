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

package view

import (
	"math"
	"strconv"

	"golang.org/x/net/context"

	"github.com/google/shenzhen-go/dev/dom"
	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
	pb "github.com/google/shenzhen-go/dev/proto/js"
)

// Graph is the view-model of a graph.
type Graph struct {
	*View
	gc GraphController

	*model.Graph
	Nodes    map[string]*Node
	Channels map[string]*Channel
}

func (g *Graph) createNode(partType string) {
	go g.reallyCreateNode(partType) // don't block in callback
}

func (g *Graph) reallyCreateNode(partType string) {
	// Invent a reasonable unique name.
	name := partType

	for i := 2; ; i++ {
		if _, found := g.Nodes[name]; !found {
			break
		}
		name = partType + " " + strconv.Itoa(i)
	}
	pt := model.PartTypes[partType].New()
	pm, err := model.MarshalPart(pt)
	if err != nil {
		g.View.Diagram.setError("Couldn't marshal part: "+err.Error(), 0, 0)
		return
	}

	n := &Node{
		View: g.View,
		Node: &model.Node{
			Name:         name,
			Enabled:      true,
			Wait:         true,
			Multiplicity: 1,
			Part:         pt,
			// TODO: use a better initial position
			X: 150,
			Y: 150,
		},
	}

	_, err = g.View.Client.CreateNode(context.Background(), &pb.CreateNodeRequest{
		Graph: g.FilePath,
		Props: &pb.NodeConfig{
			Name:         n.Name,
			Enabled:      n.Enabled,
			Wait:         n.Wait,
			Multiplicity: uint32(n.Multiplicity),
			PartType:     partType,
			PartCfg:      pm.Part,
		},
	})
	if err != nil {
		g.View.Diagram.setError("Couldn't create a new node: "+err.Error(), 0, 0)
		return
	}
	g.View.Diagram.clearError()

	n.makeElements()
	g.Nodes[name] = n
}

func (g *Graph) nearestPoint(x, y float64) (quad float64, pt Point) {
	quad = math.MaxFloat64
	test := func(p Point) {
		px, py := p.Pt()
		dx, dy := x-px, y-py
		if t := dx*dx + dy*dy; t < quad {
			quad, pt = t, p
		}
	}
	for _, n := range g.Nodes {
		for _, p := range n.AllPins {
			test(p)
		}
	}
	for _, c := range g.Channels {
		test(c)
	}
	return quad, pt
}

func (g *Graph) save(dom.Object) {
	go g.reallySave() // cannot block in callback
}

func (g *Graph) reallySave() {
	if err := g.gc.Save(context.TODO()); err != nil {
		g.View.Diagram.setError("Couldn't save: "+err.Error(), 0, 0)
	}
}

func (g *Graph) saveProperties(dom.Object) {
	go g.reallySaveProperties() // cannot block in callback
}

func (g *Graph) reallySaveProperties() {
	req := &pb.SetGraphPropertiesRequest{
		Graph:       g.FilePath,
		Name:        g.View.graphNameTextInput.Get("value").String(),
		PackagePath: g.View.graphPackagePathTextInput.Get("value").String(),
		IsCommand:   g.View.graphIsCommandCheckbox.Get("checked").Bool(),
	}
	if _, err := g.View.Client.SetGraphProperties(context.Background(), req); err != nil {
		g.View.Diagram.setError("Couldn't save: "+err.Error(), 0, 0)
		return
	}
	// And commit locally
	g.Name = req.Name
	g.PackagePath = req.PackagePath
	g.IsCommand = req.IsCommand
}

// refresh ensures the view model matches the model.
func (g *Graph) refresh() {
	// Ensure data structures are set up
	if g.Channels == nil {
		g.Channels = make(map[string]*Channel, len(g.Graph.Channels))
	}
	if g.Nodes == nil {
		g.Nodes = make(map[string]*Node, len(g.Graph.Nodes))
	}

	// Remove any channels that no longer exist.
	for k, c := range g.Channels {
		if _, found := g.Graph.Channels[k]; found {
			continue
		}
		// Remove this channel.
		c.unmakeElements()
		delete(g.Channels, k)
	}

	// Add any channels that didn't exist but now do.
	// Refresh any existing channels.
	for k, c := range g.Graph.Channels {
		if _, found := g.Channels[k]; found {
			// TODO: ch.refresh()
			continue
		}
		// Add the channel.
		ch := &Channel{
			View:    g.View,
			Channel: c,
			Pins:    make(map[*Pin]struct{}),
			created: true,
		}
		g.Channels[k] = ch
		ch.makeElements()
	}

	// Remove any nodes that no longer exist.
	for k, n := range g.Nodes {
		if _, found := g.Graph.Nodes[k]; found {
			continue
		}
		// Remove this channel.
		n.unmakeElements()
		delete(g.Nodes, k)
	}

	// Add any nodes that didn't exist but now do.
	// Refresh existing nodes.
	for k, n := range g.Graph.Nodes {
		if _, found := g.Nodes[k]; found {
			// TODO: m.refresh()
			continue
		}
		m := &Node{
			View: g.View,
			Node: n,
		}
		pd := n.Pins()
		for _, p := range pd {
			q := &Pin{
				Name:  p.Name,
				Type:  p.Type,
				input: p.Direction == pin.Input,
			}
			if q.input {
				m.Inputs = append(m.Inputs, q)
			} else {
				m.Outputs = append(m.Outputs, q)
			}
			if b := n.Connections[p.Name]; b != "" {
				if c := g.Channels[b]; c != nil {
					q.ch = c
					c.Pins[q] = struct{}{}
				}
			}
		}
		// Consolidate slices (not that it really matters)
		m.AllPins = append(m.Inputs, m.Outputs...)
		m.Inputs, m.Outputs = m.AllPins[:len(m.Inputs)], m.AllPins[len(m.Inputs):]

		g.Nodes[n.Name] = m
		m.makeElements()
	}

	// Refresh existing connections
	for _, c := range g.Channels {
		c.reposition(nil)
		c.commit()
		for p := range c.Pins {
			p.l.SetAttribute("stroke", normalColour).Show()
		}
	}
}
