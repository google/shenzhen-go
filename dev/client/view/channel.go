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

	steiner            dom.Element // symbol representing the channel itself, not used if channel is simple
	logical            Point       // centre of steiner point, for snapping
	visual             Point       // temporary centre of steiner point, for display
	dragLine, dragCirc dom.Element // temporarily visible, for dragging to more pins
	potentialPin       *Pin        // considering attaching to this pin
	subsumeInto        *Channel    // considering merging with this channel
	presubsumption     map[*Pin]struct{}
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
	c.layout(nil)
	c.logical = c.visual
	go c.reallyCommit()
}

func (c *Channel) reallyCommit() {
	if err := c.cc.Commit(context.TODO()); err != nil {
		c.errors.setError("Couldn't create a channel: " + err.Error())
		return
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
}

func (c *Channel) noSnap(x, y float64) {
	c.errors.clearError()
	c.dragTo(x, y)
	c.showDrag()
	c.SetColour(activeColour)
	c.layout(Pt(x, y))
	if c.potentialPin != nil {
		c.removePin(c.potentialPin)
		c.potentialPin.SetColour(normalColour)
		c.potentialPin = nil
	}
	if c.subsumeInto != nil {
		c.unsubsume()
	}
}

func (c *Channel) drag(x, y float64) {
	log.Print("*Channel.drag")

	d, q := c.graph.nearestPoint(x, y)

	// If the distance is too far, no snap in all cases.
	if d >= snapDist {
		c.noSnap(x, y)
		return
	}

	switch z := q.(type) {
	case *Pin:
		// Already considering connecting to this pin?
		if c.potentialPin == z {
			return
		}

		// Already subsumed into a channel?
		if c.subsumeInto != nil {
			// ...into this channel?
			if c.subsumeInto == z.channel {
				return
			}
			// Subsumed into another channel?
			c.unsubsume()
		}

		// Already connected to this pin?
		if c.hasPin(z) {
			c.noSnap(x, y)
			return
		}

		// Was considering connecting to a pin, but now connecting to a different pin.
		if c.potentialPin != nil && c.potentialPin != z {
			c.removePin(c.potentialPin)
			c.potentialPin.SetColour(normalColour)
		}

		// Trying to snap to a different channel via a pin.
		if z.channel != nil && z.channel != c {
			z.channel.subsume(c)
			return
		}

		// Snap to pin z! This means add it and hide the drag elements.
		c.errors.clearError()
		c.potentialPin = z
		c.addPin(z)
		c.SetColour(activeColour)
		c.hideDrag()
		c.layout(nil)

	case *Channel:
		// Already subsuming into this channel?
		if c.subsumeInto == z {
			return
		}

		// Connecting to itself somehow?
		if c == z {
			c.noSnap(x, y)
			return
		}

		// Was connecting to a pin before?
		if c.potentialPin != nil {
			c.removePin(c.potentialPin)
			c.potentialPin.SetColour(normalColour)
			c.potentialPin = nil
		}

		// Was subsumed into another channel?
		if c.subsumeInto != nil && c.subsumeInto != z {
			c.unsubsume()
		}

		// Trying to snap to a different channel directly.
		z.subsume(c)
	}
}

func (c *Channel) drop() {
	log.Print("*Channel.drop")

	c.errors.clearError()
	c.SetColour(normalColour)

	if c.subsumeInto != nil {
		c.subsumeInto.SetColour(normalColour)
		c.subsumeInto.commit()
	}
	if len(c.Pins) < 2 { // includes subsumption case
		go c.reallyDelete()
		return
	}
	c.potentialPin = nil
	c.commit()
	c.hideDrag()
}

func (c *Channel) addPin(p *Pin) {
	p.channel = c
	c.Pins[p].Remove()
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

func (c *Channel) subsume(ch *Channel) {
	ch.potentialPin = nil
	ch.subsumeInto = c
	ch.presubsumption = make(map[*Pin]struct{})
	for p := range ch.Pins {
		ch.presubsumption[p] = struct{}{}
		ch.removePin(p)
		c.addPin(p)
	}
	c.SetColour(activeColour)
	c.layout(nil)
	ch.SetColour(activeColour)
	ch.layout(nil)
}

func (c *Channel) unsubsume() {
	for p := range c.presubsumption {
		c.subsumeInto.removePin(p)
		c.addPin(p)
	}
	c.subsumeInto.SetColour(normalColour)
	c.subsumeInto.layout(nil)
	c.SetColour(normalColour)
	c.layout(nil)
	c.subsumeInto = nil
	c.presubsumption = nil
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
	if err := c.cc.Delete(context.TODO()); err != nil {
		c.errors.setError("Couldn't delete channel: " + err.Error())
		return
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
	c.visual /= Pt(float64(np), 0)
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
