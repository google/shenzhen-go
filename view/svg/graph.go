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
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

const (
	activeColour = "#09f"
	normalColour = "#000"

	nodeRectStyle  = "fill: #eff; stroke: #355; stroke-width:1"
	nodeTextStyle  = "font-family:Go; font-size:16; style=user-select:none; pointer-events:none"
	nodeWidth      = 160
	nodeHeight     = 50
	nodeTextOffset = 5

	pinRadius = 5
	lineWidth = 2
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
					{Name: "foo1", Type: "int"},
					{Name: "foo2", Type: "int"},
					{Name: "foo3", Type: "int"},
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
					{Name: "baz0", Type: "string"},
					{Name: "baz1", Type: "string"},
					{Name: "baz2", Type: "string"},
					{Name: "baz3", Type: "string"},
				},
				Outputs: []*Pin{
					{Name: "qux", Type: "float64"},
				},
				X: 100,
				Y: 200,
			},
		},
	}
)

func makeSVGElement(n string) *js.Object { return document.Call("createElementNS", svgNS, n) }
func cursorPos(e *js.Object) (x, y float64) {
	bcr := diagramSVG.Call("getBoundingClientRect")
	x = e.Get("clientX").Float() - bcr.Get("left").Float()
	y = e.Get("clientY").Float() - bcr.Get("top").Float()
	return
}

type Pin struct {
	Name, Type string

	x, y float64    // computed, not relative to node
	circ *js.Object // my main representation
	l, c *js.Object // line, circle, when dragging from a pin
}

func (p *Pin) mouseDown(e *js.Object) {
	currentThingy = p

	p.circ.Call("setAttribute", "fill", activeColour)

	x, y := cursorPos(e)

	// Line
	p.l = makeSVGElement("line")
	diagramSVG.Call("appendChild", p.l)
	p.l.Call("setAttribute", "x1", p.x)
	p.l.Call("setAttribute", "y1", p.y)
	p.l.Call("setAttribute", "x2", x)
	p.l.Call("setAttribute", "y2", y)
	p.l.Call("setAttribute", "stroke", activeColour)
	p.l.Call("setAttribute", "stroke-width", lineWidth)

	// Another circ
	p.c = makeSVGElement("circle")
	diagramSVG.Call("appendChild", p.c)
	p.c.Call("setAttribute", "cx", x)
	p.c.Call("setAttribute", "cy", y)
	p.c.Call("setAttribute", "r", pinRadius)
	p.c.Call("setAttribute", "fill", "transparent")
	p.c.Call("setAttribute", "stroke", activeColour)
	p.c.Call("setAttribute", "stroke-width", lineWidth)
}

func (p *Pin) dragTo(e *js.Object) {
	x, y := cursorPos(e)
	p.l.Call("setAttribute", "x2", x)
	p.l.Call("setAttribute", "y2", y)
	p.c.Call("setAttribute", "cx", x)
	p.c.Call("setAttribute", "cy", y)
}

func (p *Pin) dragComplete(e *js.Object) {
	p.circ.Call("setAttribute", "fill", normalColour)
	diagramSVG.Call("removeChild", p.l)
	diagramSVG.Call("removeChild", p.c)
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
	text.Call("setAttribute", "style", nodeTextStyle)
	text.Call("appendChild", document.Call("createTextNode", n.Name))

	// Pins
	for _, p := range n.Inputs {
		p.circ = makeSVGElement("circle")
		g.Call("appendChild", p.circ)
		p.circ.Call("setAttribute", "r", pinRadius)
		p.circ.Call("setAttribute", "fill", normalColour)
		p.circ.Call("addEventListener", "mousedown", p.mouseDown)
	}

	for _, p := range n.Outputs {
		p.circ = makeSVGElement("circle")
		g.Call("appendChild", p.circ)
		p.circ.Call("setAttribute", "r", pinRadius)
		p.circ.Call("setAttribute", "fill", normalColour)
		p.circ.Call("addEventListener", "mousedown", p.mouseDown)
	}
	n.recomputeNodePositions()

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
	n.recomputeNodePositions()
}

func (n *Node) recomputeNodePositions() {
	isp := nodeWidth / float64(len(n.Inputs)+1)
	for i, p := range n.Inputs {
		x, y := isp*float64(i+1), float64(-pinRadius)
		p.circ.Call("setAttribute", "cx", x)
		p.circ.Call("setAttribute", "cy", y)
		p.x, p.y = x+n.X, y+n.Y
	}

	osp := nodeWidth / float64(len(n.Outputs)+1)
	for i, p := range n.Outputs {
		x, y := osp*float64(i+1), float64(nodeHeight+pinRadius)
		p.circ.Call("setAttribute", "cx", x)
		p.circ.Call("setAttribute", "cy", y)
		p.x, p.y = x+n.X, y+n.Y
	}
}

type Graph struct {
	Nodes []Node
	// Channels: simple & complex
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
		t.dragComplete(e)
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
