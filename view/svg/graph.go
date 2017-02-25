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

//+build js

package main

import (
	"errors"
	"fmt"
	"math"

	"github.com/gopherjs/gopherjs/js"
)

const (
	activeColour = "#09f"
	normalColour = "#000"
	errorColour  = "#f06"

	nodeRectStyle  = "fill: #eff; stroke: #355; stroke-width:1"
	nodeTextStyle  = "font-family:Go; font-size:16; user-select:none; pointer-events:none"
	nodeWidth      = 160
	nodeHeight     = 50
	nodeTextOffset = 5

	pinRadius = 5
	lineWidth = 2
	snapQuad  = 256
)

var (
	document   = js.Global.Get("document")
	diagramSVG = document.Call("getElementById", "diagram")
	svgNS      = diagramSVG.Get("namespaceURI")

	currentThingy interface{}
	relX, relY    float64

	graph = &Graph{
		Nodes: []Node{
			{
				Name: "Hello, yes",
				Inputs: []*Pin{
					{Name: "foo1", Type: "int", input: true},
					{Name: "foo2", Type: "int", input: true},
					{Name: "foo3", Type: "string", input: true},
				},
				Outputs: []*Pin{
					{Name: "bar", Type: "string"},
					{Name: "baz", Type: "string"},
				},
				X: 100,
				Y: 100,
			},
			{
				Name: "this is dog",
				Inputs: []*Pin{
					{Name: "baz0", Type: "string", input: true},
					{Name: "baz1", Type: "string", input: true},
					{Name: "baz2", Type: "string", input: true},
					{Name: "baz3", Type: "string", input: true},
				},
				Outputs: []*Pin{
					{Name: "qux", Type: "float64"},
				},
				X: 100,
				Y: 200,
			},
		},
		Channels: make(map[*Channel]struct{}),
	}
)

func makeSVGElement(n string) *js.Object { return document.Call("createElementNS", svgNS, n) }
func cursorPos(e *js.Object) (x, y float64) {
	bcr := diagramSVG.Call("getBoundingClientRect")
	x = e.Get("clientX").Float() - bcr.Get("left").Float()
	y = e.Get("clientY").Float() - bcr.Get("top").Float()
	return
}

type Point interface {
	Pt() (x, y float64)
}

type Pin struct {
	Name, Type string

	input bool  // am I an input?
	node  *Node // owner.

	l    *js.Object // attached line; x1, y1 = x, y; x2, y2 = ch.x, ch.y.
	x, y float64    // computed, not relative to node
	circ *js.Object // my main representation
	c    *js.Object // circle, when dragging from a pin
	ch   *Channel   // attached to this channel
}

func (p *Pin) connectTo(q Point) error {
	switch q := q.(type) {
	case *Pin:
		if p.ch != nil && p.ch != q.ch {
			p.disconnect()
		}
		if q.Type != p.Type {
			return fmt.Errorf("mismatching types [%s != %s]", p.Type, q.Type)
		}
		if q.ch != nil {
			return p.connectTo(q.ch)
		}
		if p.input == q.input {
			return errors.New("both pins have the same direction")
		}
		if p.node == q.node {
			return errors.New("both pins are on the same goroutine")
		}

		// Create a new channel to connect to
		ch := &Channel{
			Type: p.Type,
			Pins: map[*Pin]struct{}{
				p: struct{}{},
				q: struct{}{},
			},
			steiner: makeSVGElement("circle"),
		}
		diagramSVG.Call("appendChild", ch.steiner)
		ch.steiner.Call("setAttribute", "r", pinRadius)
		ch.reposition()
		p.ch, q.ch = ch, ch
		graph.Channels[ch] = struct{}{}
		q.l.Call("setAttribute", "display", "")

	case *Channel:
		if p.ch != nil && p.ch != q {
			p.disconnect()
		}
		if q.Type != p.Type {
			return fmt.Errorf("mismatching types [%s != %s]", p.Type, q.Type)
		}

		// Attach to the existing channel
		p.ch = q
		q.Pins[p] = struct{}{}
		q.reposition()
	}
	return nil
}

func (p *Pin) disconnect() {
	if p.ch == nil {
		return
	}
	delete(p.ch.Pins, p)
	p.ch.reposition()
	if len(p.ch.Pins) < 2 {
		// Delete the channel
		for q := range p.ch.Pins {
			q.ch = nil
		}
		delete(graph.Channels, p.ch)
	}
	p.ch = nil
}

func (p *Pin) setPos(rx, ry float64) {
	p.circ.Call("setAttribute", "cx", rx)
	p.circ.Call("setAttribute", "cy", ry)
	p.x, p.y = rx+p.node.X, ry+p.node.Y
	if p.l != nil {
		p.l.Call("setAttribute", "x1", p.x)
		p.l.Call("setAttribute", "y1", p.y)
	}
	if p.ch != nil {
		p.ch.reposition()
	}
}

func (p *Pin) Pt() (x, y float64) { return p.x, p.y }

func (p *Pin) String() string { return fmt.Sprintf("%s.%s", p.node.Name, p.Name) }

func (p *Pin) dragStart(e *js.Object) {
	// If the pin is attached to something, don't start to drag.
	if p.ch != nil {
		return
	}
	currentThingy = p

	p.circ.Call("setAttribute", "fill", activeColour)

	x, y := cursorPos(e)
	p.l.Call("setAttribute", "x2", x)
	p.l.Call("setAttribute", "y2", y)
	p.c.Call("setAttribute", "cx", x)
	p.c.Call("setAttribute", "cy", y)
	p.c.Call("setAttribute", "stroke", activeColour)
	p.l.Call("setAttribute", "stroke", activeColour)
	p.c.Call("setAttribute", "display", "")
	p.l.Call("setAttribute", "display", "")
}

func (p *Pin) dragTo(e *js.Object) {
	x, y := cursorPos(e)
	defer func() {
		p.l.Call("setAttribute", "x2", x)
		p.l.Call("setAttribute", "y2", y)
		p.c.Call("setAttribute", "cx", x)
		p.c.Call("setAttribute", "cy", y)
	}()
	d, q := graph.nearestPoint(x, y)

	// Don't connect P to itself, snap to near the pointer, connect inputs to outputs.
	if p != q && d < snapQuad {

		if err := p.connectTo(q); err != nil {
			// TODO: complain to user via message
			p.circ.Call("setAttribute", "fill", errorColour)
			p.l.Call("setAttribute", "stroke", errorColour)
			p.c.Call("setAttribute", "stroke", errorColour)
			p.c.Call("setAttribute", "display", "")
			return
		}
		// Snap to q.ch, or q if q is a channel.
		switch q := q.(type) {
		case *Pin:
			x, y = q.ch.x, q.ch.y
		case *Channel:
			x, y = q.x, q.y
		}

		// Valid snap - ensure the colour is active.
		p.ch.setColour(activeColour)
		p.c.Call("setAttribute", "display", "none")
		return
	}

	// Nothing nearby - use active colour and unsnap if necessary.
	if p.ch != nil {
		p.ch.setColour(normalColour)
		p.disconnect()
	}

	p.circ.Call("setAttribute", "fill", activeColour)
	p.l.Call("setAttribute", "stroke", activeColour)
	p.c.Call("setAttribute", "stroke", activeColour)
	p.c.Call("setAttribute", "display", "")
}

func (p *Pin) drop(e *js.Object) {
	p.circ.Call("setAttribute", "fill", normalColour)
	p.c.Call("setAttribute", "display", "none")
	if p.ch == nil {
		p.l.Call("setAttribute", "display", "none")
		return
	}
	p.ch.setColour(normalColour)
}

func (p *Pin) makePinElement(n *Node) *js.Object {
	p.node = n
	p.circ = makeSVGElement("circle")
	p.circ.Call("setAttribute", "r", pinRadius)
	p.circ.Call("setAttribute", "fill", normalColour)
	p.circ.Call("addEventListener", "mousedown", p.dragStart)

	// Line
	p.l = makeSVGElement("line")
	diagramSVG.Call("appendChild", p.l)
	p.l.Call("setAttribute", "stroke-width", lineWidth)
	p.l.Call("setAttribute", "display", "none")

	// Another circ
	p.c = makeSVGElement("circle")
	diagramSVG.Call("appendChild", p.c)
	p.c.Call("setAttribute", "r", pinRadius)
	p.c.Call("setAttribute", "fill", "transparent")
	p.c.Call("setAttribute", "stroke-width", lineWidth)
	p.c.Call("setAttribute", "display", "none")
	return p.circ
}

type Node struct {
	Name            string
	Inputs, Outputs []*Pin
	X, Y            float64

	g *js.Object
}

func (n *Node) makeNodeElement() {
	// Group
	g := makeSVGElement("g")
	diagramSVG.Call("appendChild", g)
	g.Call("setAttribute", "transform", fmt.Sprintf("translate(%f, %f)", n.X, n.Y))

	// Rectangle
	rect := makeSVGElement("rect")
	g.Call("appendChild", rect)
	rect.Call("setAttribute", "width", nodeWidth)
	rect.Call("setAttribute", "height", nodeHeight)
	rect.Call("setAttribute", "style", nodeRectStyle)
	rect.Call("addEventListener", "mousedown", n.mouseDown)

	// Text
	text := makeSVGElement("text")
	g.Call("appendChild", text)
	text.Call("setAttribute", "x", nodeWidth/2)
	text.Call("setAttribute", "y", nodeHeight/2+nodeTextOffset)
	text.Call("setAttribute", "text-anchor", "middle")
	text.Call("setAttribute", "unselectable", "on")
	text.Call("setAttribute", "style", nodeTextStyle)
	text.Call("appendChild", document.Call("createTextNode", n.Name))

	// Pins
	for _, p := range n.Inputs {
		g.Call("appendChild", p.makePinElement(n))
	}
	for _, p := range n.Outputs {
		g.Call("appendChild", p.makePinElement(n))
	}
	n.updatePinPositions()

	// Done!
	n.g = g
}

func (n *Node) mouseDown(e *js.Object) {
	currentThingy = n
	relX, relY = e.Get("clientX").Float()-n.X, e.Get("clientY").Float()-n.Y

	// Bring to front
	diagramSVG.Call("appendChild", n.g)
}

func (n *Node) moveTo(x, y float64) {
	tf := n.g.Get("transform").Get("baseVal").Call("getItem", 0).Get("matrix")
	tf.Set("e", x)
	tf.Set("f", y)
	n.X, n.Y = x, y
	n.updatePinPositions()
}

func (n *Node) updatePinPositions() {
	isp := nodeWidth / float64(len(n.Inputs)+1)
	for i, p := range n.Inputs {
		p.setPos(isp*float64(i+1), float64(-pinRadius))
	}

	osp := nodeWidth / float64(len(n.Outputs)+1)
	for i, p := range n.Outputs {
		p.setPos(osp*float64(i+1), float64(nodeHeight+pinRadius))
	}
}

type Channel struct {
	Type string
	Cap  int

	Pins map[*Pin]struct{}

	steiner *js.Object // symbol representing the channel itself, not used if channel is simple
	x, y    float64    // centre of steiner point
}

func (c *Channel) Pt() (x, y float64) { return c.x, c.y }

func (c *Channel) reposition() {
	if len(c.Pins) < 2 {
		// Not actually a channel anymore - hide.
		c.steiner.Call("setAttribute", "display", "none")
		for t := range c.Pins {
			t.c.Call("setAttribute", "display", "none")
			t.l.Call("setAttribute", "display", "none")
		}
		return
	}
	c.x, c.y = 0, 0
	for t := range c.Pins {
		c.x += t.x
		c.y += t.y
	}
	n := float64(len(c.Pins))
	c.x /= n
	c.y /= n
	c.steiner.Call("setAttribute", "cx", c.x)
	c.steiner.Call("setAttribute", "cy", c.y)
	for t := range c.Pins {
		t.l.Call("setAttribute", "x2", c.x)
		t.l.Call("setAttribute", "y2", c.y)
	}
	disp := ""
	if len(c.Pins) == 2 {
		disp = "none"
	}
	c.steiner.Call("setAttribute", "display", disp)
}

func (c *Channel) setColour(col string) {
	c.steiner.Call("setAttribute", "fill", col)
	for t := range c.Pins {
		t.circ.Call("setAttribute", "fill", col)
		t.l.Call("setAttribute", "stroke", col)
	}
}

type Graph struct {
	Nodes []Node
	// Channels: simple & complex

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

func mouseMove(e *js.Object) {
	if currentThingy == nil {
		return
	}
	switch t := currentThingy.(type) {
	case *Node:
		t.moveTo(e.Get("clientX").Float()-relX, e.Get("clientY").Float()-relY)
	case *Pin:
		t.dragTo(e)
	}
}

func mouseUp(e *js.Object) {
	if currentThingy == nil {
		return
	}
	mouseMove(e)

	switch t := currentThingy.(type) {
	case *Node:
		// Nothing
	case *Pin:
		t.drop(e)
	}
	currentThingy = nil
}

func main() {
	for _, n := range graph.Nodes {
		n.makeNodeElement()
	}

	diagramSVG.Call("addEventListener", "mousemove", mouseMove)
	diagramSVG.Call("addEventListener", "mouseup", mouseUp)
}
