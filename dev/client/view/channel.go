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
	"golang.org/x/net/context"
)

// Channel is the view's model of a channel.
type Channel struct {
	Group // Container for all the channel elements.

	cc     ChannelController
	view   *View
	errors errorViewer
	graph  *Graph

	// Cache of raw Pin objects which are connected.
	Pins map[*Pin]*Route

	created bool // create operation sent to server?

	steiner            dom.Element // symbol representing the channel itself, not used if channel is simple
	logical            Point       // centre of steiner point, for snapping
	visual             Point       // temporary centre of steiner point, for display
	dragLine, dragCirc dom.Element // temporarily visible, for dragging to more pins
	potentialPin       *Pin        // considering attaching to this pin
}

func (c *Channel) reallyCreate() {
	if err := c.cc.Commit(context.TODO()); err != nil {
		c.errors.setError("Couldn't create a channel: " + err.Error())
		return
	}
	c.created = true
}

// MakeElements recreates elements for this channel and adds them to the parent.
func (c *Channel) MakeElements(doc dom.Document, parent dom.Element) {
	c.Group.Remove()
	c.Group = NewGroup(doc, parent)
	c.Group.Element.ClassList().Add("channel")
	c.steiner = doc.MakeSVGElement("circle").
		SetAttribute("r", pinRadius).
		AddEventListener("mousedown", c.view.dragStarter(c))
	c.steiner.ClassList().Add("draggable")
	c.dragLine = doc.MakeSVGElement("line").
		SetAttribute("stroke-width", lineWidth)
	c.dragCirc = doc.MakeSVGElement("circle").
		SetAttribute("r", pinRadius)
	c.dragCirc.ClassList().Add("draggable")

	c.hideDrag()
	c.Group.AddChildren(c.steiner, c.dragLine, c.dragCirc)
}

// Pt implements Pointer.
func (c *Channel) Pt() Point { return c.logical.Pt() }

func (c *Channel) commit() {
	if c == nil {
		return
	}
	c.logical = c.visual
	if !c.created {
		go c.reallyCreate()
	}
}

// Show the temporary drag elements.
func (c *Channel) dragTo(x, y float64) {
	c.dragLine.
		SetAttribute("x1", x).
		SetAttribute("y1", y)
	c.dragCirc.
		SetAttribute("cx", x).
		SetAttribute("cy", y)
}

func (c *Channel) showDrag() {
	c.dragLine.Show()
	c.dragCirc.Show()
	c.dragCirc.ClassList().Add("dragging")
}

func (c *Channel) hideDrag() {
	c.dragLine.Hide()
	c.dragCirc.Hide()
	c.dragCirc.ClassList().Remove("dragging")
}

func (c *Channel) dragStart(x, y float64) {
	log.Print("*Channel.dragStart")

	c.SetColour(activeColour)
	c.dragTo(x, y)
}

func (c *Channel) drag(x, y float64) {
	log.Print("*Channel.drag")

	d, q := c.graph.nearestPoint(x, y)
	p, pin := q.(*Pin)

	// Already connected to this pin?
	if pin && c.hasPin(p) && d < snapDist {
		return
	}

	noSnap := func() {
		c.dragTo(x, y)
		c.showDrag()
		c.layout(Point(complex(x, y)))
		if c.potentialPin != nil {
			c.removePin(c.potentialPin)
			c.potentialPin.SetColour(normalColour)
			c.potentialPin = nil
		}
	}

	// Was considering connecting to a pin, but now connecting to a
	// different pin or too far away?
	if pin && c.potentialPin != nil && (c.potentialPin != p || d >= snapDist) {
		noSnap()
		return
	}

	// Too far from something to snap to?
	// Trying to snap to itself?
	// Don't snap, but not an error.
	if d >= snapDist || q == c || (pin && p.channel == c) {
		c.errors.clearError()
		noSnap()
		c.SetColour(activeColour)
		return
	}

	// Trying to snap to a different channel, either directly or via a pin.
	if !pin || p.channel != nil {
		c.errors.setError("Can't connect different channels together (use another goroutine)")
		noSnap()
		c.SetColour(errorColour)
		return
	}

	// Snap to pin p!
	c.errors.clearError()
	c.potentialPin = p
	c.addPin(p)
	c.SetColour(activeColour)
	c.hideDrag()
	c.layout(nil)
}

func (c *Channel) drop() {
	log.Print("*Channel.drop")
	c.SetColour(normalColour)
	c.errors.clearError()

	if len(c.Pins) < 2 {
		go c.reallyDelete()
		return
	}
	c.potentialPin = nil
	c.layout(nil)
	c.commit()
	c.hideDrag()
}

func (c *Channel) addPin(p *Pin) {
	p.channel = c
	c.Pins[p] = NewRoute(c.view.doc, c.Group, &c.visual, p)
	c.cc.Attach(p.pc)
}

func (c *Channel) removePin(p *Pin) {
	p.channel = nil
	c.Pins[p].Remove()
	delete(c.Pins, p)
	c.cc.Detach(p.pc)
}

func (c *Channel) hasPin(p *Pin) bool {
	_, found := c.Pins[p]
	return found
}

func (c *Channel) gainFocus() {
	log.Print("TODO(josh): implement Channel.gainFocus")
}

func (c *Channel) loseFocus() {
	log.Print("TODO(josh): implement Channel.loseFocus")
}

func (c *Channel) save() {
	log.Print("TODO(josh): implement Channel.save")
}

func (c *Channel) reallyDelete() {
	if c.created {
		if err := c.cc.Delete(context.TODO()); err != nil {
			c.errors.setError("Couldn't delete channel: " + err.Error())
			return
		}
	}

	// Reset all attached pins, remove all the elements, delete from graph.
	for q := range c.Pins {
		q.channel = nil
	}
	c.Group.Remove()
	delete(c.graph.Channels, c.cc.Name())
}

func (c *Channel) layout(additional Pointer) {
	if c == nil {
		return
	}

	np := len(c.Pins)
	if additional != nil {
		np++
	}

	if np < 2 {
		// Not actually a channel anymore - hide.
		c.Hide()
		return
	}
	c.Show()

	if np < 3 {
		c.steiner.Hide()
	} else {
		c.steiner.Show()
	}

	c.visual = Point(0)
	if additional != nil {
		c.visual = additional.Pt()
	}
	for p := range c.Pins {
		c.visual += p.Pt()
	}
	c.visual /= Point(complex(float64(np), 0))
	c.steiner.
		SetAttribute("cx", real(c.visual)).
		SetAttribute("cy", imag(c.visual))
	c.dragLine.
		SetAttribute("x2", real(c.visual)).
		SetAttribute("y2", imag(c.visual))
	for _, r := range c.Pins {
		r.Reroute()
	}
}

// SetColour changes the colour of the whole channel.
func (c *Channel) SetColour(col string) {
	c.steiner.SetAttribute("fill", col)
	c.dragCirc.SetAttribute("fill", col)
	c.dragLine.SetAttribute("stroke", col)
	for p, r := range c.Pins {
		p.SetColour(col)
		r.SetStroke(col)
	}
}
