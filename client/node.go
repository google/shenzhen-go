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

// +build js

package main

import (
	"fmt"
	"log"

	"github.com/google/shenzhen-go/jsutil"
	"github.com/google/shenzhen-go/model"
	pb "github.com/google/shenzhen-go/proto"
	"github.com/gopherjs/gopherjs/js"
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

var (
	nodeMetadataSubpanel = jsutil.MustGetElement("node-metadata-panel")
	nodeCurrentSubpanel  = nodeMetadataSubpanel

	nodeNameInput         = jsutil.MustGetElement("node-name")
	nodeEnabledInput      = jsutil.MustGetElement("node-enabled")
	nodeMultiplicityInput = jsutil.MustGetElement("node-multiplicity")
	nodeWaitInput         = jsutil.MustGetElement("node-wait")

	nodePartEditors = make(map[string]*partEditor, len(model.PartTypes))
)

type partEditor struct {
	Links  *jsutil.Element
	Panels map[string]*jsutil.Element
}

func init() {
	for n, t := range model.PartTypes {
		jsutil.MustGetElement("node-new-link:"+n).
			AddEventListener("click", func(*js.Object) { theGraph.createNode(n) })
		p := make(map[string]*jsutil.Element, len(t.Panels))
		for _, d := range t.Panels {
			p[d.Name] = jsutil.MustGetElement(fmt.Sprintf("node-%s-%s-panel", n, d.Name))
		}
		nodePartEditors[n] = &partEditor{
			Links:  jsutil.MustGetElement(fmt.Sprintf("node-%s-links", n)),
			Panels: p,
		}
	}
}

// Node is the view's model of a node.
type Node struct {
	*model.Node

	Inputs, Outputs []*Pin
	X, Y            float64

	box *textBox

	relX, relY float64 // relative client offset for moving around

	subpanel *jsutil.Element // temporarily remember last subpanel for each node
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (n *Node) makeElements() {
	minWidth := nodeWidthPerPin * (max(len(n.Inputs), len(n.Outputs)) + 1)
	n.box = newTextBox(n.Name, nodeTextStyle, nodeNormalRectStyle, n.X, n.Y, float64(minWidth), nodeHeight)
	n.box.rect.AddEventListener("mousedown", n.mouseDown)
	n.box.rect.AddEventListener("mousedown", theDiagram.selecter(n))

	// Pins
	for _, p := range n.Inputs {
		n.box.group.AddChildren(p.makePinElement(n))
	}
	for _, p := range n.Outputs {
		n.box.group.AddChildren(p.makePinElement(n))
	}
	n.updatePinPositions()
}

func (n *Node) mouseDown(e *js.Object) {
	theDiagram.dragItem = n
	n.relX, n.relY = e.Get("clientX").Float()-n.X, e.Get("clientY").Float()-n.Y

	// Bring to front
	theDiagram.AddChildren(n.box.group)
}

func (n *Node) drag(e *js.Object) {
	x, y := e.Get("clientX").Float()-n.relX, e.Get("clientY").Float()-n.relY
	n.box.moveTo(x, y)
	n.X, n.Y = x, y
	n.updatePinPositions()
}

func (n *Node) drop(e *js.Object) {
	x, y := e.Get("clientX").Float()-n.relX, e.Get("clientY").Float()-n.relY
	req := &pb.SetPositionRequest{
		Graph: graphPath,
		Node:  n.Name,
		X:     int64(x),
		Y:     int64(y),
	}

	go func() { // cannot block in callback
		if _, err := theClient.SetPosition(context.Background(), req); err != nil {
			log.Printf("Couldn't SetPosition: %v", err)
		}
	}()
}

type focusable interface {
	GainFocus(*js.Object)
}

func (n *Node) gainFocus(e *js.Object) {
	n.box.rect.SetAttribute("style", nodeSelectedRectStyle)
	nodeNameInput.Set("value", n.Node.Name)
	nodeEnabledInput.Set("checked", n.Node.Enabled)
	nodeMultiplicityInput.Set("value", n.Node.Multiplicity)
	nodeWaitInput.Set("checked", n.Node.Wait)
	showRHSPanel(nodePropertiesPanel)
	nodePartEditors[n.Node.Part.TypeKey()].Links.Get("style").Set("display", "inline")
	if n.subpanel == nil {
		n.subpanel = nodeMetadataSubpanel
	}
	n.showSubPanel(n.subpanel)
	if f := n.Node.Part.(focusable); f != nil {
		f.GainFocus(e)
	}
}

func (n *Node) loseFocus(e *js.Object) {
	n.box.rect.SetAttribute("style", nodeNormalRectStyle)
}

func (n *Node) save(e *js.Object) {
	pj, err := model.MarshalPart(n.Part)
	if err != nil {
		log.Printf("Couldn't marshal part: %v", err)
		return
	}
	props := &pb.NodeConfig{
		Name:         nodeNameInput.Get("value").String(),
		Enabled:      nodeEnabledInput.Get("checked").Bool(),
		Multiplicity: uint32(nodeMultiplicityInput.Get("value").Int()),
		Wait:         nodeWaitInput.Get("checked").Bool(),
		PartCfg:      pj.Part,
		PartType:     pj.Type,
		X:            int64(n.X),
		Y:            int64(n.Y),
	}
	req := &pb.SetNodePropertiesRequest{
		Graph: graphPath,
		Node:  n.Node.Name,
		Props: props,
	}
	go func() {
		if _, err := theClient.SetNodeProperties(context.Background(), req); err != nil {
			log.Printf("Couldn't update node properties: %v", err)
			return
		}
		// Update local copy, since these were read at save time.
		if n.Name != props.Name {
			delete(theGraph.Nodes, n.Name)
			theGraph.Nodes[props.Name] = n
			n.Name = props.Name // TODO: simplify view-model
			n.Node.Name = props.Name

			n.box.setText(props.Name)
			n.updatePinPositions()
		}
		n.Node.Enabled = props.Enabled
		n.Node.Multiplicity = uint(props.Multiplicity)
		n.Node.Wait = props.Wait
	}()
}

func (n *Node) delete(*js.Object) {
	log.Print("TODO(josh): implement Node.delete")
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

func (n *Node) showSubPanel(p *jsutil.Element) {
	n.subpanel = p
	if nodeCurrentSubpanel == p {
		return
	}
	nodeCurrentSubpanel.Get("style").Set("display", "none")
	nodeCurrentSubpanel = p
	nodeCurrentSubpanel.Get("style").Set("display", nil)
}
