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
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/google/shenzhen-go/api"
	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/model/pin"
	"github.com/gopherjs/gopherjs/js"
)

// Graph is the view's model of a graph.
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

func (g *Graph) saveProperties(*js.Object) {
	req := &api.SetGraphPropertiesRequest{
		Request: api.Request{
			Graph: graphPath,
		},
		Name:        document.Call("getElementById", "graph-prop-name").Get("value").String(),
		PackagePath: document.Call("getElementById", "graph-prop-package-path").Get("value").String(),
		IsCommand:   document.Call("getElementById", "graph-prop-is-command").Get("checked").Bool(),
	}
	go func() { // cannot block in callback
		if err := client.SetGraphProperties(req); err != nil {
			log.Printf("Couldn't SetGraphProperties: %v", err)
		}
	}()
}

func loadGraph(d *diagram) (*Graph, error) {

	gj := js.Global.Get("GraphJSON")
	if gj == nil {
		return nil, errors.New("no global GraphJSON")
	}
	g, err := model.LoadJSON(strings.NewReader(gj.String()), "", "")
	if err != nil {
		return nil, fmt.Errorf("decoding GraphJSON: %v", err)
	}

	graph := new(Graph)
	chans := make(map[string]*Channel)
	graph.Channels = make(map[*Channel]struct{})
	for k, c := range g.Channels {
		ch := &Channel{
			Type: c.Type,
			Cap:  c.Capacity,
			Pins: make(map[*Pin]struct{}),
			d:    d,
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
			d:    d,
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
				if c := chans[b]; c != nil {
					q.ch = c
					c.Pins[q] = struct{}{}
				}
			}
		}
		graph.Nodes[n.Name] = m
	}
	return graph, nil
}
