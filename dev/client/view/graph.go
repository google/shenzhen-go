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

	"golang.org/x/net/context"

	"github.com/google/shenzhen-go/dev/dom"
	"github.com/google/shenzhen-go/dev/model/pin"
)

// Graph is the view-model of a graph.
type Graph struct {
	view *View
	gc   GraphController

	Nodes    map[string]*Node
	Channels map[string]*Channel
}

func (g *Graph) createNode(partType string) {
	go g.reallyCreateNode(partType) // don't block in callback
}

func (g *Graph) reallyCreateNode(partType string) {
	nc, err := g.gc.CreateNode(context.TODO(), partType)
	if err != nil {
		g.view.diagram.setError("Couldn't create a new node: "+err.Error(), 0, 0)
		return
	}
	g.view.diagram.clearError()

	n := &Node{
		view: g.view,
		nc:   nc,
	}
	n.makeElements()
	g.Nodes[nc.Node().Name] = n
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
		g.view.diagram.setError("Couldn't save: "+err.Error(), 0, 0)
	}
}

func (g *Graph) saveProperties(dom.Object) {
	go g.reallySaveProperties() // cannot block in callback
}

func (g *Graph) reallySaveProperties() {
	if err := g.gc.SaveProperties(context.TODO()); err != nil {
		g.view.diagram.setError("Couldn't save properties: "+err.Error(), 0, 0)
	}
}

// refresh ensures the view model matches the model.
func (g *Graph) refresh() {
	// Ensure data structures are set up
	if g.Channels == nil {
		g.Channels = make(map[string]*Channel, g.gc.NumChannels())
	}
	if g.Nodes == nil {
		g.Nodes = make(map[string]*Node, g.gc.NumNodes())
	}

	// Remove any channels that no longer exist.
	for k, c := range g.Channels {
		if g.gc.Channel(k) != nil {
			continue
		}
		// Remove this channel.
		c.unmakeElements()
		delete(g.Channels, k)
	}

	// Add any channels that didn't exist but now do.
	// Refresh any existing channels.
	//for k, c := range g.gc.Graph().Channels {
	g.gc.Channels(func(cc ChannelController) {
		k := cc.Channel().Name
		if g.Channels[k] != nil {
			// TODO: ch.refresh()
			return
		}
		// Add the channel.
		ch := &Channel{
			view:    g.view,
			channel: cc.Channel(),
			Pins:    make(map[*Pin]struct{}),
			created: true,
		}
		g.Channels[k] = ch
		ch.makeElements()
	})

	// Remove any nodes that no longer exist.
	for k, n := range g.Nodes {
		if g.gc.Node(k) != nil {
			continue
		}
		// Remove this channel.
		n.unmakeElements()
		delete(g.Nodes, k)
	}

	// Add any nodes that didn't exist but now do.
	// Refresh existing nodes.
	//for k, n := range g.gc.Graph().Nodes {
	g.gc.Nodes(func(nc NodeController) {
		k := nc.Node().Name
		if g.Nodes[k] != nil {
			// TODO: m.refresh()
			return
		}
		m := &Node{
			view: g.view,
			nc:   nc,
		}
		pd := nc.Node().Pins()
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
			if b := nc.Node().Connections[p.Name]; b != "" {
				if c := g.Channels[b]; c != nil {
					q.ch = c
					c.Pins[q] = struct{}{}
				}
			}
		}
		// Consolidate slices (not that it really matters)
		m.AllPins = append(m.Inputs, m.Outputs...)
		m.Inputs, m.Outputs = m.AllPins[:len(m.Inputs)], m.AllPins[len(m.Inputs):]

		g.Nodes[nc.Node().Name] = m
		m.makeElements()
	})

	// Refresh existing connections
	for _, c := range g.Channels {
		c.reposition(nil)
		c.commit()
		for p := range c.Pins {
			p.l.SetAttribute("stroke", normalColour).Show()
		}
	}
}
