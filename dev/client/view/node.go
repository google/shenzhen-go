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

	"github.com/google/shenzhen-go/dev/dom"
	"golang.org/x/net/context"
)

const (
	nodeNormalRectStyle   = "fill: #eff; stroke: #355; stroke-width:1"
	nodeSelectedRectStyle = "fill: #cef; stroke: #145; stroke-width:2"
	nodeTextStyle         = "font-family:Go; font-size:16; user-select:none; pointer-events:none"
	nodeWidthPerPin       = 20
	nodeHeight            = 50
	nodeBoxMargin         = 20
	nodeTextOffsetY       = 5
)

// Node is the view's model of a node.
type Node struct {
	Group
	TextBox *TextBox
	Inputs  []*Pin
	Outputs []*Pin
	AllPins []*Pin

	nc     NodeController
	view   *View
	errors errorViewer
	graph  *Graph

	relX, relY float64 // relative client offset for moving around
	x, y       float64 // cache of actual position
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MakeElements makes the elements that are part of this node.
func (n *Node) MakeElements(doc dom.Document, parent dom.Element) *Node {
	n.Group = NewGroup(doc, parent).MoveTo(n.nc.Position())

	minWidth := nodeWidthPerPin * (max(len(n.Inputs), len(n.Outputs)) + 1)
	n.TextBox = &TextBox{
		Margin:      nodeBoxMargin,
		TextOffsetY: nodeTextOffsetY,
		MinWidth:    float64(minWidth),
	}
	n.TextBox.
		MakeElements(doc, n.Group).
		SetTextStyle(nodeTextStyle).
		SetRectStyle(nodeNormalRectStyle).
		SetHeight(nodeHeight).
		SetText(n.nc.Name()).
		RecomputeWidth()

	n.TextBox.Rect.
		AddEventListener("mousedown", n.view.dragStarter(n)).
		AddEventListener("mousedown", n.view.selecter(n))

	// Pins
	for _, p := range n.AllPins {
		p.MakeElements(doc, n.Group)
		p.node = n
	}
	n.updatePinPositions()
	return n
}

// AddTo adds the node as a child of the given parent.
func (n *Node) AddTo(parent dom.Element) *Node {
	parent.AddChildren(n.Group)
	return n
}

// Remove removes the node from the group's parent.
func (n *Node) Remove() {
	n.Group.Parent().RemoveChildren(n.Group)
}

// MoveTo moves the textbox to have the topleft corner at x, y.
func (n *Node) MoveTo(x, y float64) *Node {
	n.Group.SetAttribute("transform", fmt.Sprintf("translate(%f, %f)", x, y))
	n.x, n.y = x, y
	return n
}

func (n *Node) dragStart(x, y float64) {
	n.relX, n.relY = x-n.x, y-n.y

	// Bring to front
	n.Group.Parent().AddChildren(n.Group)
}

func (n *Node) drag(x, y float64) {
	x, y = x-n.relX, y-n.relY
	n.MoveTo(x, y)
	n.updatePinPositions()
}

func (n *Node) drop() {
	go func() { // cannot block in callback
		if err := n.nc.SetPosition(context.TODO(), n.x, n.y); err != nil {
			n.errors.setError("Couldn't set the position: " + err.Error())
		}
	}()
}

func (n *Node) gainFocus() {
	n.TextBox.Rect.SetAttribute("style", nodeSelectedRectStyle)
	n.nc.GainFocus()
}

func (n *Node) loseFocus() {
	n.TextBox.Rect.SetAttribute("style", nodeNormalRectStyle)
	n.nc.LoseFocus()
}

func (n *Node) save() {
	go n.reallySave()
}

func (n *Node) reallySave() {
	oldName := n.nc.Name()
	if err := n.nc.Save(context.TODO()); err != nil {
		n.errors.setError("Couldn't update node properties: " + err.Error())
		return
	}
	// Update local copy, since these were read at save time.
	// TODO: refresh pins
	if name := n.nc.Name(); name != oldName {
		delete(n.graph.Nodes, oldName)
		n.graph.Nodes[name] = n
		n.TextBox.SetText(name)
		n.updatePinPositions()
	}
}

func (n *Node) delete() {
	go n.reallyDelete() // don't block handler
}

func (n *Node) reallyDelete() {
	// Chatty, but cleans everything up.
	for _, p := range n.AllPins {
		p.reallyDisconnect()
	}
	if err := n.nc.Delete(context.TODO()); err != nil {
		n.errors.setError("Couldn't delete: " + err.Error())
		return
	}
	delete(n.graph.Nodes, n.nc.Name())
	n.Remove()
}

func (n *Node) refresh() {
	// TODO: implement
	n.updatePinPositions()
}

func (n *Node) updatePinPositions() {
	// Pins have to be aware of both their global and local coordinates,
	// so the nearest one can be found, and channels can be drawn correctly.
	w := n.TextBox.Width()
	isp := w / float64(len(n.Inputs)+1)
	for i, p := range n.Inputs {
		p.MoveTo(isp*float64(i+1), float64(-pinRadius))
	}

	osp := w / float64(len(n.Outputs)+1)
	for i, p := range n.Outputs {
		p.MoveTo(osp*float64(i+1), float64(nodeHeight+pinRadius))
	}
}
