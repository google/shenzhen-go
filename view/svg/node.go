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
	"log"

	"github.com/google/shenzhen-go/api"
	"github.com/gopherjs/gopherjs/js"
)

const (
	nodeRectStyle   = "fill: #eff; stroke: #355; stroke-width:1"
	nodeTextStyle   = "font-family:Go; font-size:16; user-select:none; pointer-events:none"
	nodeWidthPerPin = 20
	nodeHeight      = 50
	nodeTextOffset  = 5
)

// Node is the view's model of a node.
type Node struct {
	Name            string
	Inputs, Outputs []*Pin
	X, Y            float64

	d   *diagram
	box *textBox

	relX, relY float64 // relative client offset for moving around
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (n *Node) makeElements() {
	minWidth := nodeWidthPerPin * (max(len(n.Inputs), len(n.Outputs)) + 1)
	n.box = newTextBox(n.d, n.Name, nodeTextStyle, nodeRectStyle, n.X, n.Y, float64(minWidth), nodeHeight)
	n.box.rect.Call("addEventListener", "mousedown", n.mouseDown)

	// Pins
	for _, p := range n.Inputs {
		n.box.group.Call("appendChild", p.makePinElement(n))
	}
	for _, p := range n.Outputs {
		n.box.group.Call("appendChild", p.makePinElement(n))
	}
	n.updatePinPositions()
}

func (n *Node) mouseDown(e *js.Object) {
	n.d.dragItem = n
	n.relX, n.relY = e.Get("clientX").Float()-n.X, e.Get("clientY").Float()-n.Y

	// Bring to front
	n.d.Call("appendChild", n.box.group)
}

func (n *Node) drag(e *js.Object) {
	x, y := e.Get("clientX").Float()-n.relX, e.Get("clientY").Float()-n.relY
	n.box.moveTo(x, y)
	n.X, n.Y = x, y
	n.updatePinPositions()
}

func (n *Node) drop(e *js.Object) {
	x, y := e.Get("clientX").Float()-n.relX, e.Get("clientY").Float()-n.relY
	req := &api.SetPositionRequest{
		NodeRequest: api.NodeRequest{
			Request: api.Request{
				Graph: graphPath,
			},
			Node: n.Name,
		},
		X: int(x),
		Y: int(y),
	}

	go func() { // cannot block in callback
		if err := client.SetPosition(req); err != nil {
			log.Printf("Couldn't SetPosition: %v", err)
		}
	}()
}

func (n *Node) updatePinPositions() {
	// Pins have to be aware of both their global and local coordinates,
	// so the nearest one can be found, and channels can be drawn correctly.
	isp := n.box.width / float64(len(n.Inputs)+1)
	for i, p := range n.Inputs {
		p.setPos(isp*float64(i+1), float64(-pinRadius))
	}

	osp := n.box.width / float64(len(n.Outputs)+1)
	for i, p := range n.Outputs {
		p.setPos(osp*float64(i+1), float64(nodeHeight+pinRadius))
	}
}
