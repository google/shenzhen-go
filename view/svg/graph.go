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

package main

import (
	"encoding/json"
	"log"
	"math"
	"strings"

	"github.com/google/shenzhen-go/api"

	"github.com/gopherjs/gopherjs/js"
)

type Graph struct {
	Nodes    map[string]*Node
	Channels map[*Channel]struct{}
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
		for _, p := range n.Inputs {
			test(p)
		}
		for _, p := range n.Outputs {
			test(p)
		}
	}
	for c := range g.Channels {
		test(c)
	}
	return quad, pt
}

func loadGraph() {
	graph = new(Graph)

	gj := js.Global.Get("GraphJSON")
	if gj == nil {
		return
	}
	d := json.NewDecoder(strings.NewReader(gj.String()))
	var g api.Graph
	if err := d.Decode(&g); err != nil {
		log.Printf("Decoding GraphJSON: %v", err)
		return
	}

	chans := make(map[string]*Channel)
	graph.Channels = make(map[*Channel]struct{})
	for k, c := range g.Channels {
		ch := &Channel{
			Type: c.Type,
			Cap:  c.Capacity,
			Pins: make(map[*Pin]struct{}),
		}
		chans[k] = ch
		graph.Channels[ch] = struct{}{}
	}

	graph.Nodes = make(map[string]*Node, len(g.Nodes))
	for _, n := range g.Nodes {
		m := &Node{
			Name: n.Name,
			X:    float64(n.X),
			Y:    float64(n.Y),
		}
		for k, p := range n.Pins {
			q := &Pin{
				Name:  k,
				Type:  p.Type,
				input: p.Direction == api.Input,
			}
			if q.input {
				m.Inputs = append(m.Inputs, q)
			} else {
				m.Outputs = append(m.Outputs, q)
			}

			if c, ok := chans[p.Binding]; ok {
				q.ch = c
				c.Pins[q] = struct{}{}
			}
		}
		graph.Nodes[n.Name] = m
	}

}
