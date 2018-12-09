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
	"sort"
	"syscall/js"

	"github.com/google/shenzhen-go/dom"
)

// Pin represents a node pin visually, and has enough information to know
// if it is validly connected.
type Pin struct {
	Group             // Container for all the pin elements.
	Shape dom.Element // The pin itself.

	// Computed, absolute coordinates (not relative to node).
	point Point

	pc PinController

	view    *View
	errors  errorViewer
	graph   *Graph
	node    *Node    // owner.
	channel *Channel // attached to this channel, is often nil
}

// MoveTo moves the pin (relatively).
func (p *Pin) MoveTo(rel Point) {
	p.Group.MoveTo(rel)
	p.point = rel + p.node.abs
	if p.channel != nil {
		p.channel.layout(nil)
		p.channel.logical = p.channel.visual
	}
}

// Pt returns the diagram coordinate of the pin, for nearest-neighbor purposes.
func (p *Pin) Pt() Point { return p.point }

func (p *Pin) String() string { return p.node.nc.Name() + "." + p.pc.Name() }

func (p *Pin) dragStart(pt Point) {
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
	ch.dragStart(pt)
}

func (p *Pin) hoverText() string { return p.pc.Name() + " (" + p.pc.Type() + ")" }

func (p *Pin) mouseEnter(e js.Value) {
	p.view.showHoverTip(e, p.hoverText())
}

func (p *Pin) mouseLeave(js.Value) {
	p.view.hoverTip.Hide()
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
		AddEventListener("mousedown", p.view.selecter(p)).
		AddEventListener("mouseenter", js.NewEventCallback(js.StopPropagation, p.mouseEnter)).
		AddEventListener("mouseleave", js.NewEventCallback(0, p.mouseLeave))

	p.Shape.ClassList().Add("draggable")

	p.Group.AddChildren(p.Shape)
	p.unselected()
	return p
}

func (p *Pin) selected() {
	p.Group.ClassList().Add("selected")
}

func (p *Pin) unselected() {
	p.Group.ClassList().Remove("selected")
}

func (p *Pin) gainFocus() {
	if p.channel == nil {
		return
	}
	p.view.changeSelection(p.channel)
}

func (p *Pin) loseFocus() {
	// Nop.
}

func sortPins(ps []*Pin) {
	sort.Slice(ps, func(i, j int) bool {
		pi, pj := ps[i].pc, ps[j].pc
		if pi.IsInput() == pj.IsInput() {
			return pi.Name() < pj.Name()
		}
		return pi.IsInput()
	})
}
