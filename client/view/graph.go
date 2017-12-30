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
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/model/pin"
	pb "github.com/google/shenzhen-go/proto/js"
	"github.com/gopherjs/gopherjs/js"
	"golang.org/x/net/context"
)

var (
	graphNameElement        = theDocument.ElementByID("graph-prop-name")
	graphPackagePathElement = theDocument.ElementByID("graph-prop-package-path")
	graphIsCommandElement   = theDocument.ElementByID("graph-prop-is-command")
)

// Graph is the view-model of a graph.
type Graph struct {
	*model.Graph
	Nodes    map[string]*Node
	Channels map[string]*Channel
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
				X:            150,
				Y:            150,
			},
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
		for _, p := range n.AllPins {
			test(p)
		}
	}
	for _, c := range g.Channels {
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
	for k := range g.Channels {
		if _, found := g.Graph.Channels[k]; found {
			continue
		}
		// Remove this channel.
		// TODO: c.removeElements()
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
			Channel: c,
			Pins:    make(map[*Pin]struct{}),
		}
		g.Channels[k] = ch
		ch.makeElements()
	}

	// Remove any nodes that no longer exist.
	for k := range g.Nodes {
		if _, found := g.Graph.Nodes[k]; found {
			continue
		}
		// Remove this channel.
		// TODO: n.removeElements()
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
			p.l.SetAttribute("stroke", normalColour)
			p.l.Show()
		}
	}
}

func loadGraph(gj string) (*Graph, error) {
	g, err := model.LoadJSON(strings.NewReader(gj), "", "")
	if err != nil {
		return nil, fmt.Errorf("decoding GraphJSON: %v", err)
	}

	graph := &Graph{Graph: g}
	graph.refresh()
	return graph, nil
}
