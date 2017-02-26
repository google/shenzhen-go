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
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

const (
	nodeRectStyle  = "fill: #eff; stroke: #355; stroke-width:1"
	nodeTextStyle  = "font-family:Go; font-size:16; user-select:none; pointer-events:none"
	nodeWidth      = 160
	nodeHeight     = 50
	nodeTextOffset = 5
)

type Node struct {
	Name            string
	Inputs, Outputs []*Pin
	X, Y            float64

	g *js.Object

	relX, relY float64 // relative client offset for moving around
}

func (n *Node) makeElements() {
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
	dragItem = n
	n.relX, n.relY = e.Get("clientX").Float()-n.X, e.Get("clientY").Float()-n.Y

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
