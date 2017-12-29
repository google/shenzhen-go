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

// +build js

package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/google/shenzhen-go/jsutil"
	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/model/pin"
	pb "github.com/google/shenzhen-go/proto"
	"github.com/gopherjs/gopherjs/js"
	"golang.org/x/net/context"
)

var (
	graphNameElement        = jsutil.MustGetElement("graph-prop-name")
	graphPackagePathElement = jsutil.MustGetElement("graph-prop-package-path")
	graphIsCommandElement   = jsutil.MustGetElement("graph-prop-is-command")
)

// Graph is the view's model of a graph.
type Graph struct {
	Nodes    map[string]*Node
	Channels map[*Channel]struct{}
}

func (g *Graph) createNode(partType string) {
	go func() {
		// Invent a reasonable unique name.
		name := partType
		i := 2
		for {
			if _, found := g.Nodes[name]; !found {
				break
			}
			name = partType + " " + strconv.Itoa(i)
			i++
		}
		pt := model.PartTypes[partType].New()
		pm, err := model.MarshalPart(pt)
		if err != nil {
			log.Printf("Couldn't marshal (brand new!) part: %v", err)
			return
		}

		n := &Node{
			Node: &model.Node{
				Name:         name,
				Enabled:      true,
				Wait:         true,
				Multiplicity: 1,
				Part:         pt,
			},
			X: 150,
			Y: 150,
		}

		_, err = theClient.CreateNode(context.Background(), &pb.CreateNodeRequest{
			Graph: graphPath,
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
			log.Printf("Couldn't CreateNode: %v", err)
			return
		}

		n.makeElements()
		g.Nodes[name] = n
	}()
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

func (g *Graph) save(*js.Object) {
	go func() { // cannot block in callback
		if _, err := theClient.Save(context.Background(), &pb.SaveRequest{Graph: graphPath}); err != nil {
			log.Printf("Couldn't Save: %v", err)
		}
	}()
}

func (g *Graph) saveProperties(*js.Object) {
	go func() { // cannot block in callback
		req := &pb.SetGraphPropertiesRequest{
			Graph:       graphPath,
			Name:        graphNameElement.Get("value").String(),
			PackagePath: graphPackagePathElement.Get("value").String(),
			IsCommand:   graphIsCommandElement.Get("checked").Bool(),
		}
		if _, err := theClient.SetGraphProperties(context.Background(), req); err != nil {
			log.Printf("Couldn't SetGraphProperties: %v", err)
		}
	}()
}

func loadGraph(gj string) (*Graph, error) {
	g, err := model.LoadJSON(strings.NewReader(gj), "", "")
	if err != nil {
		return nil, fmt.Errorf("decoding GraphJSON: %v", err)
	}

	graph := new(Graph)
	chans := make(map[string]*Channel)
	graph.Channels = make(map[*Channel]struct{})
	for k, c := range g.Channels {
		ch := &Channel{
			Channel: c,
			Pins:    make(map[*Pin]struct{}),
		}
		chans[k] = ch
		graph.Channels[ch] = struct{}{}
		ch.makeElements()
	}

	graph.Nodes = make(map[string]*Node, len(g.Nodes))
	for _, n := range g.Nodes {
		m := &Node{
			Node: n,
			X:    float64(n.X),
			Y:    float64(n.Y),
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
		m.makeElements()
	}
	// Show existing connections
	for c := range graph.Channels {
		c.reposition(nil)
		c.commit()
		for p := range c.Pins {
			p.l.SetAttribute("stroke", normalColour)
			p.l.Show()
		}
	}
	return graph, nil
}
