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
	"context"
	"log"
	"math"
	"math/cmplx"

	"github.com/google/shenzhen-go/dev/dom"
)

// Graph is the view-model of a graph.
type Graph struct {
	Group // container for all graph elements
	gc    GraphController

	doc    dom.Document // responsible for creating new elements dynamically
	view   *View        // for passing to
	errors errorViewer

	Nodes    map[string]*Node
	Channels map[string]*Channel
}

func (g *Graph) reallyCreateNode(partType string) {
	nc, err := g.gc.CreateNode(context.TODO(), partType)
	if err != nil {
		g.errors.setError("Couldn't create a new node: " + err.Error())
		return
	}
	g.errors.clearError()

	n := &Node{
		view:   g.view,
		errors: g.errors,
		graph:  g,
		nc:     nc,
	}
	n.abs = Pt(nc.Position())
	n.MakeElements(g.doc, g.Group)
	g.Nodes[nc.Name()] = n
}

func (g *Graph) nearestPoint(x, y float64) (dist float64, pt Pointer) {
	dist = math.MaxFloat64
	q := complex(x, y)
	test := func(p Pointer) {
		if t := cmplx.Abs(q - C(p)); t < dist {
			dist, pt = t, p
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
	return dist, pt
}

// goroutines because cannot block in callback
func (g *Graph) save(dom.Object)     { go g.reallySave() }
func (g *Graph) revert(dom.Object)   { go g.reallyRevert() }
func (g *Graph) generate(dom.Object) { go g.reallyGenerate() }
func (g *Graph) build(dom.Object)    { go g.reallyBuild() }
func (g *Graph) install(dom.Object)  { go g.reallyInstall() }
func (g *Graph) run(dom.Object)      { go g.reallyRun() }

func (g *Graph) reallySave() {
	if err := g.gc.Save(context.TODO()); err != nil {
		g.errors.setError("Couldn't save: " + err.Error())
	}
}

func (g *Graph) reallyRevert() {
	if err := g.gc.Revert(context.TODO()); err != nil {
		g.errors.setError("Couldn't revert: " + err.Error())
	}
}

func (g *Graph) reallyGenerate() {
	if err := g.gc.Generate(context.TODO()); err != nil {
		g.errors.setError("Couldn't generate: " + err.Error())
	}
}

func (g *Graph) reallyBuild() {
	if err := g.gc.Build(context.TODO()); err != nil {
		g.errors.setError("Couldn't build: " + err.Error())
	}
}

func (g *Graph) reallyInstall() {
	if err := g.gc.Install(context.TODO()); err != nil {
		g.errors.setError("Couldn't install: " + err.Error())
	}
}

func (g *Graph) reallyRun() {
	if err := g.gc.Run(context.TODO()); err != nil {
		g.errors.setError("Couldn't run: " + err.Error())
	}
}

func (g *Graph) commit(dom.Object) {
	go g.reallyCommit() // cannot block in callback
}

func (g *Graph) reallyCommit() {
	if err := g.gc.Commit(context.TODO()); err != nil {
		g.errors.setError("Couldn't save properties: " + err.Error())
	}
}

func (g *Graph) gainFocus() { g.gc.GainFocus() }
func (g *Graph) loseFocus() { go g.reallyCommit() }

// MakeElements drops any existing elements, and then loads new ones
// from the graph controller.
func (g *Graph) MakeElements(doc dom.Document, parent dom.Element) {
	g.Group.Remove()
	g.Group = NewGroup(doc, parent)

	// Set up data structures.
	g.Channels = make(map[string]*Channel, g.gc.NumChannels())
	g.Nodes = make(map[string]*Node, g.gc.NumNodes())

	// Add any channels that didn't exist but now do.
	// Refresh any existing channels.
	g.gc.Channels(func(cc ChannelController) {
		// Add the channel.
		ch := &Channel{
			view:   g.view,
			errors: g.errors,
			graph:  g,
			cc:     cc,
			Pins:   make(map[*Pin]*Route),
		}
		g.Channels[cc.Name()] = ch
		ch.MakeElements(doc, g.Group)
	})

	// Add any nodes that didn't exist but now do.
	// Refresh existing nodes.
	g.gc.Nodes(func(nc NodeController) {
		m := &Node{
			view:   g.view,
			errors: g.errors,
			graph:  g,
			nc:     nc,
			abs:    Pt(nc.Position()),
		}
		inputs := 0
		nc.Pins(func(pc PinController, channel string) {
			q := &Pin{
				pc:     pc,
				view:   g.view,
				errors: g.errors,
				graph:  g,
				node:   m,
			}
			if pc.IsInput() {
				inputs++
			}
			m.AllPins = append(m.AllPins, q)
			if channel == "" || channel == "nil" {
				return
			}
			c := g.Channels[channel]
			if c == nil {
				log.Printf("channel %q not found", channel)
				return
			}
			c.addPin(q)
		})
		// Consolidate slices (not that it really matters)
		sortPins(m.AllPins)
		m.Inputs, m.Outputs = m.AllPins[:inputs], m.AllPins[inputs:]

		g.Nodes[nc.Name()] = m
		m.MakeElements(doc, g.Group)
	})

	// Load connections.
	for _, ch := range g.Channels {
		ch.layout(nil)
		ch.logical = ch.visual
		ch.Show()
		for _, r := range ch.Pins {
			r.Reroute()
		}
	}
}
