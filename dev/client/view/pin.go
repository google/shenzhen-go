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
	"log"

	"github.com/google/shenzhen-go/dev/dom"
)

// Pin represents a node pin visually, and has enough information to know
// if it is validly connected.
type Pin struct {
	Group               // Container for all the pin elements.
	Shape   dom.Element // The pin itself.
	Nametag *TextBox    // Temporarily visible on hover.
	//dragLine, dragCirc dom.Element // Temporary elements when dragging from unattached pin.

	// Computed, absolute coordinates (not relative to node).
	x, y float64

	pc PinController

	view    *View
	errors  errorViewer
	graph   *Graph
	node    *Node    // owner.
	channel *Channel // attached to this channel, is often nil
}

// MoveTo moves the pin (relatively).
func (p *Pin) MoveTo(rx, ry float64) {
	p.Group.MoveTo(rx, ry)
	p.x, p.y = rx+p.node.x, ry+p.node.y
	p.channel.layout(nil)
	p.channel.commit()
}

// Pt returns the diagram coordinate of the pin, for nearest-neighbor purposes.
func (p *Pin) Pt() (x, y float64) { return p.x, p.y }

func (p *Pin) String() string { return p.node.nc.Name() + "." + p.pc.Name() }

func (p *Pin) dragStart(x, y float64) {
	log.Print("*Pin.dragStart")

	ch := p.channel
	if ch == nil {
		if err := p.view.createChannel(p); err != nil {
			p.errors.setError("Couldn't create channel: " + err.Error())
			return
		}
		ch = p.channel
	} else {
		ch.potentialPin = p
	}
	p.view.dragItem = ch
	ch.dragStart(x, y)
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

// MakeElements recreates elements associated with this pin.
func (p *Pin) MakeElements(doc dom.Document, parent dom.Element) *Pin {
	// Container for the pin elements.
	p.Group.Remove()
	p.Group = NewGroup(doc, parent)
	p.Group.Element.ClassList().Add("pin")

	// The pin itself, visually.
	p.Shape = doc.MakeSVGElement("circle").
		SetAttribute("r", pinRadius).
		AddEventListener("mousedown", p.view.dragStarter(p)).
		AddEventListener("mouseenter", p.mouseEnter).
		AddEventListener("mouseleave", p.mouseLeave)

	p.Shape.ClassList().Add("draggable")

	// Nametag textbox.
	p.Nametag = &TextBox{Margin: 20}
	p.Nametag.
		MakeElements(doc, p.Group).
		SetHeight(30).
		SetText(p.pc.Name() + " (" + p.pc.Type() + ")")
	p.Nametag.RecomputeWidth()
	p.Nametag.Hide()

	p.Group.AddChildren(p.Shape)
	p.SetColour(normalColour)
	return p
}

// SetColour sets the colour of the pin (and dragging elements).
func (p *Pin) SetColour(colour string) {
	p.Shape.SetAttribute("fill", colour)
}
