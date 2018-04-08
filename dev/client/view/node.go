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
	nodeBoxMargin         = 20
	nodeTextOffsetY       = 5
)

// Node is the view's model of a node.
type Node struct {
	*TextBox

	nc   NodeController
	view *View

	Inputs, Outputs, AllPins []*Pin

	relX, relY float64 // relative client offset for moving around

	subpanel dom.Element // temporarily remember last subpanel for each node
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MakeElements makes the elements that are part of this node.
func (n *Node) MakeElements(doc dom.Document) {
	minWidth := nodeWidthPerPin * (max(len(n.Inputs), len(n.Outputs)) + 1)
	n.TextBox = (&TextBox{
		Margin:      nodeBoxMargin,
		TextOffsetY: nodeTextOffsetY,
		MinWidth:    float64(minWidth),
	}).
		MakeElements(doc).
		SetText(n.nc.Name()).
		SetTextStyle(nodeTextStyle).
		SetRectangleStyle(nodeNormalRectStyle).
		SetHeight(nodeHeight).
		MoveTo(n.nc.Position())
	n.view.diagram.AddChildren(n.TextBox)
	n.TextBox.Rectangle.
		AddEventListener("mousedown", n.mouseDown).
		AddEventListener("mousedown", n.view.diagram.selecter(n))

	// Pins
	for _, p := range n.AllPins {
		n.TextBox.AddChildren(p.makeElements(n))
	}
	n.updatePinPositions()
}

func (n *Node) Remove() {
	n.TextBox.Remove()
}

func (n *Node) mouseDown(e dom.Object) {
	n.view.diagram.dragItem = n
	nx, ny := n.nc.Position()
	n.relX, n.relY = e.Get("clientX").Float()-nx, e.Get("clientY").Float()-ny

	// Bring to front
	n.TextBox.Parent().AddChildren(n.TextBox)
}

func (n *Node) drag(e dom.Object) {
	x, y := e.Get("clientX").Float()-n.relX, e.Get("clientY").Float()-n.relY
	n.TextBox.MoveTo(x, y)
	n.nc.Node().X, n.nc.Node().Y = x, y
	n.updatePinPositions()
}

func (n *Node) drop(e dom.Object) {
	go func() { // cannot block in callback
		x, y := e.Get("clientX").Float()-n.relX, e.Get("clientY").Float()-n.relY
		req := &pb.SetPositionRequest{
			Graph: n.view.graph.gc.Graph().FilePath,
			Node:  n.nc.Node().Name,
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
	n.TextBox.Rectangle.SetAttribute("style", nodeSelectedRectStyle)
	n.view.nodeNameInput.Set("value", n.nc.Node().Name)
	n.view.nodeEnabledInput.Set("checked", n.nc.Node().Enabled)
	n.view.nodeMultiplicityInput.Set("value", n.nc.Node().Multiplicity)
	n.view.nodeWaitInput.Set("checked", n.nc.Node().Wait)
	n.view.ShowRHSPanel(n.view.NodePropertiesPanel)
	n.view.nodePartEditors[n.nc.Node().Part.TypeKey()].Links.Show()
	if n.subpanel == nil {
		n.subpanel = n.view.nodeMetadataSubpanel
	}
	n.showSubPanel(n.subpanel)
	if f := n.nc.Node().Part.(focusable); f != nil {
		f.GainFocus(e)
	}
}

func (n *Node) loseFocus(e dom.Object) {
	n.TextBox.Rectangle.SetAttribute("style", nodeNormalRectStyle)
}

func (n *Node) save(e dom.Object) {
	go n.reallySave(e)
}

func (n *Node) reallySave(e dom.Object) {
	pj, err := model.MarshalPart(n.nc.Node().Part)
	if err != nil {
		n.view.diagram.setError("Couldn't marshal part: "+err.Error(), 0, 0)
		return
	}
	nx, ny := n.nc.Position()
	props := &pb.NodeConfig{
		Name:         n.view.nodeNameInput.Get("value").String(),
		Enabled:      n.view.nodeEnabledInput.Get("checked").Bool(),
		Multiplicity: uint32(n.view.nodeMultiplicityInput.Get("value").Int()),
		Wait:         n.view.nodeWaitInput.Get("checked").Bool(),
		PartCfg:      pj.Part,
		PartType:     pj.Type,
		X:            nx,
		Y:            ny,
	}
	req := &pb.SetNodePropertiesRequest{
		Graph: n.view.graph.gc.Graph().FilePath,
		Node:  n.nc.Name(),
		Props: props,
	}
	if _, err := n.view.client.SetNodeProperties(context.Background(), req); err != nil {
		n.view.diagram.setError("Couldn't update node properties: "+err.Error(), 0, 0)
		return
	}
	// Update local copy, since these were read at save time.
	// TODO: check whether the available pins have changed.
	if n.nc.Node().Name != props.Name {
		delete(n.view.graph.Nodes, n.nc.Node().Name)
		n.view.graph.Nodes[props.Name] = n
		n.nc.Node().Name = props.Name

		n.TextBox.SetText(props.Name)
		n.updatePinPositions()
	}
	n.nc.Node().Enabled = props.Enabled
	n.nc.Node().Multiplicity = uint(props.Multiplicity)
	n.nc.Node().Wait = props.Wait
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
		Node:  n.nc.Node().Name,
	}
	if _, err := n.view.client.DeleteNode(context.Background(), req); err != nil {
		n.view.diagram.setError("Couldn't delete: "+err.Error(), 0, 0)
		return
	}
	delete(n.view.graph.Nodes, n.nc.Node().Name)
	n.Remove()
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
	w := n.TextBox.Width()
	isp := w / float64(len(n.Inputs)+1)
	for i, p := range n.Inputs {
		p.setPos(isp*float64(i+1), float64(-pinRadius))
	}

	osp := w / float64(len(n.Outputs)+1)
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
