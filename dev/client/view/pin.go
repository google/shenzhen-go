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
	"golang.org/x/net/context"
)

const (
	nametagRectStyle = "fill: #efe; stroke: #353; stroke-width:1"
	nametagTextStyle = "font-family:Go; font-size:16; user-select:none; pointer-events:none"
)

// Pin represents a node pin visually, and has enough information to know
// if it is validly connected.
type Pin struct {
	Group                          // Container for all the pin elements.
	Shape              dom.Element // The pin itself.
	Nametag            *TextBox    // Temporarily visible on hover.
	dragLine, dragCirc dom.Element // Temporary elements when dragging from unattached pin.

	// Computed, absolute coordinates (not relative to node).
	x, y float64

	pc PinController

	node *Node    // owner.
	ch   *Channel // attached to this channel, is often nil
}

func (p *Pin) reallyConnect() {
	// Attach to the existing channel
	if err := p.pc.Attach(context.TODO(), p.ch.cc); err != nil {
		p.node.view.setError("Couldn't connect: " + err.Error())
	}
}

func (p *Pin) disconnect() {
	if p.ch == nil {
		return
	}
	go p.reallyDisconnect()
	delete(p.ch.Pins, p)
	p.ch.setColour(normalColour)
	p.ch.reposition(nil)
	if len(p.ch.Pins) < 2 {
		// Delete the channel
		for q := range p.ch.Pins {
			q.ch = nil
		}
		delete(p.node.view.graph.Channels, p.ch.cc.Channel().Name)
	}
	p.ch = nil
}

func (p *Pin) reallyDisconnect() {
	if err := p.pc.Detach(context.TODO()); err != nil {
		p.node.view.setError("Couldn't disconnect: " + err.Error())
	}
}

// MoveTo moves the pin (relatively).
func (p *Pin) MoveTo(rx, ry float64) {
	p.Group.MoveTo(rx, ry)
	p.x, p.y = rx+p.node.nc.Node().X, ry+p.node.nc.Node().Y
}

// Pt returns the diagram coordinate of the pin, for nearest-neighbor purposes.
func (p *Pin) Pt() (x, y float64) { return p.x, p.y }

func (p *Pin) String() string { return p.node.nc.Node().Name + "." + p.pc.Name() }

func (p *Pin) connectTo(q Point) {
	switch q := q.(type) {
	case *Pin:
		if p.ch != nil && p.ch != q.ch {
			p.disconnect()
		}
		if q.ch != nil {
			p.connectTo(q.ch)
			return
		}

		// Create a new channel to connect to
		ch := p.node.view.createChannel(p, q)
		ch.reposition(nil)
		p.node.view.graph.Channels[ch.cc.Channel().Name] = ch

	case *Channel:
		if p.ch != nil && p.ch != q {
			p.disconnect()
		}

		p.ch = q
		q.Pins[p] = &Route{}
		q.reposition(nil)
	}
	return
}

func (p *Pin) dragStart(e dom.Object) {
	// If a channel is attached, detach and drag from that instead.
	if p.ch != nil {
		p.disconnect()
		p.ch.dragStart(e)
		return
	}

	// Not attached, so the pin is the drag item and show the temporary line and circle.
	p.node.view.dragItem = p
	x, y := p.node.view.diagramCursorPos(e)

	// Start with errorColour because we're probably only in range of ourself.
	p.dragTo(x, y, errorColour)
}

func (p *Pin) drag(e dom.Object) {
	x, y := p.node.view.diagramCursorPos(e)
	colour := activeColour

	d, q := p.node.view.graph.nearestPoint(x, y)

	// Don't connect P to itself, don't connect if nearest is far away.
	if p == q || d >= snapQuad {
		p.node.view.clearError()
		if p.ch != nil {
			p.ch.setColour(normalColour)
			p.disconnect()
		}
		colour = errorColour
		p.Shape.SetAttribute("fill", colour)
		p.dragTo(x-p.x, y-p.y, colour)
		return
	}

	// Make the connection - this is the responsibility of the channel.
	p.node.view.clearError()
	colour = activeColour
	p.connectTo(q)
	p.ch.setColour(colour)
	p.hideDrag()
}

func (p *Pin) drop(e dom.Object) {
	p.node.view.clearError()
	p.Shape.SetAttribute("fill", normalColour)
	p.hideDrag()
	if p.ch == nil {
		go p.reallyDisconnect()
		return
	}
	if p.ch.created {
		go p.reallyConnect()
	}
	p.ch.setColour(normalColour)
	p.ch.commit()
}

// Show the temporary drag elements with a specific colour.
// Coordinates are pin relative.
func (p *Pin) dragTo(rx, ry float64, stroke string) {
	p.dragLine.
		SetAttribute("x2", rx).
		SetAttribute("y2", ry).
		SetAttribute("stroke", stroke).
		SetAttribute("stroke-width", lineWidth).
		Show()
	p.dragCirc.
		SetAttribute("cx", rx).
		SetAttribute("cy", ry).
		SetAttribute("stroke", stroke).
		SetAttribute("stroke-width", lineWidth).
		Show()
}

func (p *Pin) hideDrag() {
	p.dragLine.Hide()
	p.dragCirc.Hide()
}

func (p *Pin) mouseEnter(dom.Object) {
	x, y := p.x-p.node.nc.Node().X, p.y-p.node.nc.Node().Y
	if p.pc.IsInput() {
		y -= 38
	} else {
		y += 8
	}
	p.Nametag.MoveTo(x, y).Show()
}

func (p *Pin) mouseLeave(dom.Object) {
	p.Nametag.Hide()
}

// MakeElements creates elements associated with this pin.
func (p *Pin) MakeElements(doc dom.Document, parent dom.Element) *Pin {
	// Container for the pin elements.
	p.Group = NewGroup(doc, parent)

	// The pin itself, visually
	p.Shape = doc.MakeSVGElement("circle").
		SetAttribute("r", pinRadius).
		SetAttribute("fill", normalColour).
		AddEventListener("mousedown", p.dragStart).
		AddEventListener("mouseenter", p.mouseEnter).
		AddEventListener("mouseleave", p.mouseLeave)

	// Nametag textbox.
	p.Nametag = &TextBox{Margin: 20, TextOffsetY: 5}
	p.Nametag.
		MakeElements(doc, p.Group).
		SetHeight(30).
		SetTextStyle(nametagTextStyle).
		SetRectStyle(nametagRectStyle).
		SetText(p.pc.Name() + " (" + p.pc.Type() + ")")
	p.Nametag.RecomputeWidth()
	p.Nametag.Hide()

	// Temporarily-visible elements when dragging from an unattached pin.
	p.dragLine = doc.MakeSVGElement("line").
		SetAttribute("stroke", normalColour).
		Hide()
	p.dragCirc = doc.MakeSVGElement("circ").
		SetAttribute("r", pinRadius).
		SetAttribute("stroke", normalColour).
		Hide()

	p.Group.AddChildren(p.Shape, p.dragLine, p.dragCirc)
	return p
}

// AddTo adds the pin's group as a child to the given parent.
func (p *Pin) AddTo(parent dom.Element) *Pin {
	parent.AddChildren(p.Group)
	return p
}

// Remove removes the group from its parent.
func (p *Pin) Remove() {
	p.Group.Parent().RemoveChildren(p.Group)
}
