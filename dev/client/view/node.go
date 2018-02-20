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
	"github.com/google/shenzhen-go/dev/jsutil"
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
	*model.Node
	*View

	Inputs, Outputs, AllPins []*Pin

	box *textBox

	relX, relY float64 // relative client offset for moving around

	subpanel jsutil.Element // temporarily remember last subpanel for each node
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (n *Node) makeElements() {
	minWidth := nodeWidthPerPin * (max(len(n.Inputs), len(n.Outputs)) + 1)
	n.box = n.View.newTextBox(n.Name, nodeTextStyle, nodeNormalRectStyle, n.X, n.Y, float64(minWidth), nodeHeight)
	n.View.Diagram.AddChildren(n.box)
	n.box.rect.
		AddEventListener("mousedown", n.mouseDown).
		AddEventListener("mousedown", n.View.Diagram.selecter(n))

	// Pins
	for _, p := range n.AllPins {
		n.box.AddChildren(p.makeElements(n))
	}
	n.updatePinPositions()
}

func (n *Node) mouseDown(e jsutil.Object) {
	n.View.Diagram.dragItem = n
	n.relX, n.relY = e.Get("clientX").Float()-n.X, e.Get("clientY").Float()-n.Y

	// Bring to front
	n.View.Diagram.AddChildren(n.box)
}

func (n *Node) drag(e jsutil.Object) {
	x, y := e.Get("clientX").Float()-n.relX, e.Get("clientY").Float()-n.relY
	n.box.moveTo(x, y)
	n.X, n.Y = x, y
	n.updatePinPositions()
}

func (n *Node) drop(e jsutil.Object) {
	go func() { // cannot block in callback
		x, y := e.Get("clientX").Float()-n.relX, e.Get("clientY").Float()-n.relY
		req := &pb.SetPositionRequest{
			Graph: n.View.Graph.FilePath,
			Node:  n.Name,
			X:     x,
			Y:     y,
		}
		if _, err := n.View.Client.SetPosition(context.Background(), req); err != nil {
			n.View.Diagram.setError("Couldn't set the position: "+err.Error(), x, y)
		}
	}()
}

type focusable interface {
	GainFocus(jsutil.Object)
}

func (n *Node) gainFocus(e jsutil.Object) {
	n.box.rect.SetAttribute("style", nodeSelectedRectStyle)
	n.View.nodeNameInput.Set("value", n.Node.Name)
	n.View.nodeEnabledInput.Set("checked", n.Node.Enabled)
	n.View.nodeMultiplicityInput.Set("value", n.Node.Multiplicity)
	n.View.nodeWaitInput.Set("checked", n.Node.Wait)
	n.View.ShowRHSPanel(n.View.NodePropertiesPanel)
	n.View.nodePartEditors[n.Node.Part.TypeKey()].Links.Show()
	if n.subpanel == nil {
		n.subpanel = n.View.nodeMetadataSubpanel
	}
	n.showSubPanel(n.subpanel)
	if f := n.Node.Part.(focusable); f != nil {
		f.GainFocus(e)
	}
}

func (n *Node) loseFocus(e jsutil.Object) {
	n.box.rect.SetAttribute("style", nodeNormalRectStyle)
}

func (n *Node) save(e jsutil.Object) {
	go n.reallySave(e)
}

func (n *Node) reallySave(e jsutil.Object) {
	pj, err := model.MarshalPart(n.Part)
	if err != nil {
		n.View.Diagram.setError("Couldn't marshal part: "+err.Error(), 0, 0)
		return
	}
	props := &pb.NodeConfig{
		Name:         n.View.nodeNameInput.Get("value").String(),
		Enabled:      n.View.nodeEnabledInput.Get("checked").Bool(),
		Multiplicity: uint32(n.View.nodeMultiplicityInput.Get("value").Int()),
		Wait:         n.View.nodeWaitInput.Get("checked").Bool(),
		PartCfg:      pj.Part,
		PartType:     pj.Type,
		X:            n.X,
		Y:            n.Y,
	}
	req := &pb.SetNodePropertiesRequest{
		Graph: n.View.Graph.FilePath,
		Node:  n.Node.Name,
		Props: props,
	}
	if _, err := n.View.Client.SetNodeProperties(context.Background(), req); err != nil {
		n.View.Diagram.setError("Couldn't update node properties: "+err.Error(), 0, 0)
		return
	}
	// Update local copy, since these were read at save time.
	// TODO: check whether the available pins have changed.
	if n.Name != props.Name {
		delete(n.View.Graph.Nodes, n.Name)
		n.View.Graph.Nodes[props.Name] = n
		n.Name = props.Name // TODO: simplify view-model
		n.Node.Name = props.Name

		n.box.setText(props.Name)
		n.updatePinPositions()
	}
	n.Node.Enabled = props.Enabled
	n.Node.Multiplicity = uint(props.Multiplicity)
	n.Node.Wait = props.Wait
}

func (n *Node) delete(jsutil.Object) {
	go n.reallyDelete() // don't block handler
}

func (n *Node) reallyDelete() {
	// Chatty, but cleans everything up.
	for _, p := range n.AllPins {
		p.reallyDisconnect()
		p.l.Hide()
	}
	req := &pb.DeleteNodeRequest{
		Graph: n.View.Graph.FilePath,
		Node:  n.Node.Name,
	}
	if _, err := n.View.Client.DeleteNode(context.Background(), req); err != nil {
		n.View.Diagram.setError("Couldn't delete: "+err.Error(), 0, 0)
		return
	}
	delete(n.View.Graph.Nodes, n.Name)
	n.View.Diagram.RemoveChildren(n.box)
	for _, p := range n.AllPins {
		p.unmakeElements()
	}
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

func (n *Node) showSubPanel(p jsutil.Element) {
	n.subpanel = p
	if n.View.nodeCurrentSubpanel == p {
		return
	}
	n.View.nodeCurrentSubpanel.Hide()
	n.View.nodeCurrentSubpanel = p.Show()
}
