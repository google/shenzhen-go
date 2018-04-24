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

	view    *View
	errors  errorViewer
	graph   *Graph
	node    *Node    // owner.
	channel *Channel // attached to this channel, is often nil
}

func (p *Pin) reallyConnect() {
	// Attach to the existing channel
	if err := p.pc.Attach(context.TODO(), p.channel.cc); err != nil {
		p.errors.setError("Couldn't connect: " + err.Error())
	}
}

func (p *Pin) disconnect() {
	if p.channel == nil {
		return
	}
	go p.reallyDisconnect() // Don't block handler.
}

func (p *Pin) reallyDisconnect() {
	if err := p.pc.Detach(context.TODO()); err != nil {
		p.errors.setError("Couldn't disconnect: " + err.Error())
		return
	}
	if p.channel == nil {
		return
	}
	p.channel.Pins[p].Remove()
	delete(p.channel.Pins, p)
	if len(p.channel.Pins) < 2 {
		p.channel.delete()
		return
	}
	p.channel.SetColour(normalColour)
	p.channel.reposition(nil)
	p.channel = nil

}

// MoveTo moves the pin (relatively).
func (p *Pin) MoveTo(rx, ry float64) {
	p.Group.MoveTo(rx, ry)
	p.x, p.y = rx+p.node.x, ry+p.node.y
	p.channel.reposition(nil)
	p.channel.commit()
}

// Pt returns the diagram coordinate of the pin, for nearest-neighbor purposes.
func (p *Pin) Pt() (x, y float64) { return p.x, p.y }

func (p *Pin) String() string { return p.node.nc.Name() + "." + p.pc.Name() }

func (p *Pin) connectTo(q Pointer) {
	switch q := q.(type) {
	case *Pin:
		if p.channel != nil && p.channel != q.channel {
			p.disconnect()
		}

		// If the target pin is already connected, connect to that channel.
		if q.channel != nil {
			p.connectTo(q.channel)
			return
		}

		// Create a new channel to connect to.
		if err := p.view.createChannel(p, q); err != nil {
			p.errors.setError("Couldn't create channel: " + err.Error())
			return
		}

	case *Channel:
		if p.channel == q {
			return
		}

		if p.channel != nil && p.channel != q {
			p.disconnect()
		}

		p.channel = q
		q.Pins[p] = NewRoute(p.view.doc, q.Group, &q.visual, p)
		q.reposition(nil)
	}
}

func (p *Pin) dragStart(float64, float64) {
	p.SetColour(activeColour)
}

func (p *Pin) drag(x, y float64) {
	d, q := p.graph.nearestPoint(x, y)

	// Don't connect P to itself, don't connect if nearest is far away.
	if p == q || d >= snapQuad {
		p.errors.clearError()
		if p.channel != nil {
			p.channel.SetColour(normalColour)
			p.disconnect()
		}
		p.SetColour(errorColour)
		p.dragTo(x, y)
		return
	}

	// Make the connection.
	p.errors.clearError()
	p.connectTo(q)
	p.channel.SetColour(activeColour)
	p.hideDrag()
}

func (p *Pin) drop() {
	p.errors.clearError()
	p.SetColour(normalColour)
	p.hideDrag()
	if p.channel == nil {
		go p.reallyDisconnect()
		return
	}
	if p.channel.created {
		go p.reallyConnect()
	}
	p.channel.SetColour(normalColour)
	p.channel.commit()
}

// Show the temporary drag elements with a specific colour.
// Coordinates are pin relative.
func (p *Pin) dragTo(x, y float64) {
	p.dragLine.
		SetAttribute("x2", x-p.x).
		SetAttribute("y2", y-p.y).
		Show()
	p.dragCirc.
		SetAttribute("cx", x-p.x).
		SetAttribute("cy", y-p.y).
		Show()
}

func (p *Pin) hideDrag() {
	p.dragLine.Hide()
	p.dragCirc.Hide()
}

func (p *Pin) mouseEnter(dom.Object) {
	x, y := 8.0, 8.0
	if p.pc.IsInput() {
		y = -38
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

	// The pin itself, visually.
	p.Shape = doc.MakeSVGElement("circle").
		SetAttribute("r", pinRadius).
		AddEventListener("mousedown", p.view.dragStarter(p)).
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
		SetAttribute("stroke-width", lineWidth).
		Hide()
	p.dragCirc = doc.MakeSVGElement("circle").
		SetAttribute("r", pinRadius).
		Hide()

	p.Group.AddChildren(p.Shape, p.dragLine, p.dragCirc)
	p.SetColour(normalColour)
	return p
}

// SetColour sets the colour of the pin (and dragging elements).
func (p *Pin) SetColour(colour string) {
	p.Shape.SetAttribute("fill", colour)
	p.dragLine.SetAttribute("stroke", colour)
	p.dragCirc.SetAttribute("fill", colour)
}
