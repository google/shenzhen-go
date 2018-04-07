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
	"github.com/google/shenzhen-go/dev/dom"
	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/js"
	"golang.org/x/net/context"
)

const (
	nodeNormalRectStyle   = "fill: #eff; stroke: #355; stroke-width:1"
	nodeSelectedRectStyle = "fill: #cef; stroke: #145; stroke-width:2"
	nodeTextStyle         = "font-family:Go; font-size:16; user-select:none; pointer-events:none"
	nodeWidthPerPin       = 20
	nodeHeight            = 50
	nodeTextOffset        = 5
)

// Node is the view's model of a node.
type Node struct {
	node *model.Node
	view *View

	Inputs, Outputs, AllPins []*Pin

	box *textBox

	relX, relY float64 // relative client offset for moving around

	subpanel dom.Element // temporarily remember last subpanel for each node
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (n *Node) makeElements() {
	minWidth := nodeWidthPerPin * (max(len(n.Inputs), len(n.Outputs)) + 1)
	n.box = n.view.newTextBox(n.node.Name, nodeTextStyle, nodeNormalRectStyle, n.node.X, n.node.Y, float64(minWidth), nodeHeight)
	n.view.diagram.AddChildren(n.box)
	n.box.rect.
		AddEventListener("mousedown", n.mouseDown).
		AddEventListener("mousedown", n.view.diagram.selecter(n))

	// Pins
	for _, p := range n.AllPins {
		n.box.AddChildren(p.makeElements(n))
	}
	n.updatePinPositions()
}

func (n *Node) unmakeElements() {
	n.box.unmakeElements()
	n.box = nil
}

func (n *Node) mouseDown(e dom.Object) {
	n.view.diagram.dragItem = n
	n.relX, n.relY = e.Get("clientX").Float()-n.node.X, e.Get("clientY").Float()-n.node.Y

	// Bring to front
	n.view.diagram.AddChildren(n.box)
}

func (n *Node) drag(e dom.Object) {
	x, y := e.Get("clientX").Float()-n.relX, e.Get("clientY").Float()-n.relY
	n.box.moveTo(x, y)
	n.node.X, n.node.Y = x, y
	n.updatePinPositions()
}

func (n *Node) drop(e dom.Object) {
	go func() { // cannot block in callback
		x, y := e.Get("clientX").Float()-n.relX, e.Get("clientY").Float()-n.relY
		req := &pb.SetPositionRequest{
			Graph: n.view.graph.gc.Graph().FilePath,
			Node:  n.node.Name,
			X:     x,
			Y:     y,
		}
		if _, err := n.view.client.SetPosition(context.Background(), req); err != nil {
			n.view.diagram.setError("Couldn't set the position: "+err.Error(), x, y)
		}
	}()
}

type focusable interface {
	GainFocus(dom.Object)
}

func (n *Node) gainFocus(e dom.Object) {
	n.box.rect.SetAttribute("style", nodeSelectedRectStyle)
	n.view.nodeNameInput.Set("value", n.node.Name)
	n.view.nodeEnabledInput.Set("checked", n.node.Enabled)
	n.view.nodeMultiplicityInput.Set("value", n.node.Multiplicity)
	n.view.nodeWaitInput.Set("checked", n.node.Wait)
	n.view.ShowRHSPanel(n.view.NodePropertiesPanel)
	n.view.nodePartEditors[n.node.Part.TypeKey()].Links.Show()
	if n.subpanel == nil {
		n.subpanel = n.view.nodeMetadataSubpanel
	}
	n.showSubPanel(n.subpanel)
	if f := n.node.Part.(focusable); f != nil {
		f.GainFocus(e)
	}
}

func (n *Node) loseFocus(e dom.Object) {
	n.box.rect.SetAttribute("style", nodeNormalRectStyle)
}

func (n *Node) save(e dom.Object) {
	go n.reallySave(e)
}

func (n *Node) reallySave(e dom.Object) {
	pj, err := model.MarshalPart(n.node.Part)
	if err != nil {
		n.view.diagram.setError("Couldn't marshal part: "+err.Error(), 0, 0)
		return
	}
	props := &pb.NodeConfig{
		Name:         n.view.nodeNameInput.Get("value").String(),
		Enabled:      n.view.nodeEnabledInput.Get("checked").Bool(),
		Multiplicity: uint32(n.view.nodeMultiplicityInput.Get("value").Int()),
		Wait:         n.view.nodeWaitInput.Get("checked").Bool(),
		PartCfg:      pj.Part,
		PartType:     pj.Type,
		X:            n.node.X,
		Y:            n.node.Y,
	}
	req := &pb.SetNodePropertiesRequest{
		Graph: n.view.graph.gc.Graph().FilePath,
		Node:  n.node.Name,
		Props: props,
	}
	if _, err := n.view.client.SetNodeProperties(context.Background(), req); err != nil {
		n.view.diagram.setError("Couldn't update node properties: "+err.Error(), 0, 0)
		return
	}
	// Update local copy, since these were read at save time.
	// TODO: check whether the available pins have changed.
	if n.node.Name != props.Name {
		delete(n.view.graph.Nodes, n.node.Name)
		n.view.graph.Nodes[props.Name] = n
		n.node.Name = props.Name

		n.box.setText(props.Name)
		n.updatePinPositions()
	}
	n.node.Enabled = props.Enabled
	n.node.Multiplicity = uint(props.Multiplicity)
	n.node.Wait = props.Wait
}

func (n *Node) delete(dom.Object) {
	go n.reallyDelete() // don't block handler
}

func (n *Node) reallyDelete() {
	// Chatty, but cleans everything up.
	for _, p := range n.AllPins {
		p.reallyDisconnect()
		p.l.Hide()
	}
	req := &pb.DeleteNodeRequest{
		Graph: n.view.graph.gc.Graph().FilePath,
		Node:  n.node.Name,
	}
	if _, err := n.view.client.DeleteNode(context.Background(), req); err != nil {
		n.view.diagram.setError("Couldn't delete: "+err.Error(), 0, 0)
		return
	}
	delete(n.view.graph.Nodes, n.node.Name)
	n.view.diagram.RemoveChildren(n.box)
	for _, p := range n.AllPins {
		p.unmakeElements()
	}
}

func (n *Node) refresh() {
	// TODO: implement
	n.updatePinPositions()
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

func (n *Node) showSubPanel(p dom.Element) {
	n.subpanel = p
	if n.view.nodeCurrentSubpanel == p {
		return
	}
	n.view.nodeCurrentSubpanel.Hide()
	n.view.nodeCurrentSubpanel = p.Show()
}
