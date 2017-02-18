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
)

var (
	document   = js.Global.Get("document")
	diagramSVG = document.Call("getElementById", "diagram")
	svgNS      = diagramSVG.Get("namespaceURI")

	currentNode *Node
	relX, relY  float64

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

type Pin struct {
	Name, Type string

	x, y float64 // computed, relative to node
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
	for i, p := range n.Inputs {
		sp := nodeWidth / float64(len(n.Inputs)+1)
		p.x = sp * float64(i+1)
		p.y = -pinRadius
		circ := makeSVGElement("circle")
		g.Call("appendChild", circ)
		circ.Call("setAttribute", "cx", p.x)
		circ.Call("setAttribute", "cy", p.y)
		circ.Call("setAttribute", "r", pinRadius)
		circ.Call("setAttribute", "fill", normalColour)
	}

	for i, p := range n.Outputs {
		sp := nodeWidth / float64(len(n.Outputs)+1)
		p.x = sp * float64(i+1)
		p.y = nodeHeight + pinRadius
		circ := makeSVGElement("circle")
		g.Call("appendChild", circ)
		circ.Call("setAttribute", "cx", p.x)
		circ.Call("setAttribute", "cy", p.y)
		circ.Call("setAttribute", "r", pinRadius)
		circ.Call("setAttribute", "fill", normalColour)
	}

	// Done!
	n.g = g
}

func (n *Node) mouseDown(e *js.Object) {
	currentNode = n
	relX, relY = e.Get("clientX").Float()-n.X, e.Get("clientY").Float()-n.Y

	// Bring to front
	diagramSVG.Call("appendChild", n.g)
}

func (n *Node) moveTo(x, y float64) {
	tf := n.g.Get("transform").Get("baseVal").Call("getItem", 0).Get("matrix")
	tf.Set("e", x)
	tf.Set("f", y)
	n.X, n.Y = x, y
}

type Graph struct {
	Nodes []Node
	// Channels: simple & complex
}

func main() {
	for _, n := range graph.Nodes {
		n.makeNodeElement()
	}

	diagramSVG.Call("addEventListener", "mousemove", func(e *js.Object) {
		if currentNode == nil {
			return
		}
		currentNode.moveTo(e.Get("clientX").Float()-relX, e.Get("clientY").Float()-relY)
	})

	diagramSVG.Call("addEventListener", "mouseup", func(e *js.Object) {
		if currentNode == nil {
			return
		}
		currentNode.moveTo(e.Get("clientX").Float()-relX, e.Get("clientY").Float()-relY)
		currentNode = nil
	})
}
