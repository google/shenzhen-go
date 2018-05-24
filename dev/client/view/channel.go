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

	// This part works so well it's scary.
	c.Group.Element.
		AddEventListener("mousedown", c.view.selecter(c)).
		AddEventListener("mouseenter", c.mouseEnter).
		AddEventListener("mouseleave", c.mouseLeave)

	c.steiner = doc.MakeSVGElement("circle").
		SetAttribute("r", pinRadius).
		AddEventListener("mousedown", c.view.dragStarter(c)).
		AddEventListener("mousedown", c.view.selecter(c))
	c.steiner.ClassList().Add("draggable")

	c.dragLine = doc.MakeSVGElement("line")
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
		c.errors.setError("Couldn't commit a channel: " + err.Error())
		return
	}
}

func (c *Channel) mouseEnter(e dom.Object) {
	log.Print("*Channel.mouseEnter")
	c.view.showHoverTip(e, c.cc.Name())
}

func (c *Channel) mouseLeave(dom.Object) {
	c.view.hoverTip.Hide()
}

// Show the temporary drag elements.
func (c *Channel) dragTo(pt Point) {
	c.dragLine.
		SetAttribute("x1", real(pt)).
		SetAttribute("y1", imag(pt))
	c.dragCirc.
		SetAttribute("cx", real(pt)).
		SetAttribute("cy", imag(pt))
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

func (c *Channel) dragStart(p Point) {
	log.Print("*Channel.dragStart")
}

func (c *Channel) noSnap(pt Point) {
	c.errors.clearError()
	c.dragTo(pt)
	c.showDrag()
	c.layout(pt)
	if c.potentialPin != nil {
		c.removePin(c.potentialPin)
	}
	if c.subsumeInto != nil {
		c.unsubsume()
	}
}

func (c *Channel) drag(pt Point) {
	log.Print("*Channel.drag")

	d, q := c.graph.nearestPoint(real(pt), imag(pt))

	// If the distance is too far, no snap in all cases.
	if d >= snapDist {
		c.noSnap(pt)
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
			c.noSnap(pt)
			return
		}

		// Was considering connecting to a pin, but now connecting to a different pin.
		if c.potentialPin != nil && c.potentialPin != z {
			c.removePin(c.potentialPin)
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
		z.selected()
		c.hideDrag()
		c.layout(nil)

	case *Channel:
		// Already subsuming into this channel?
		if c.subsumeInto == z {
			return
		}

		// Connecting to itself somehow?
		if c == z {
			c.noSnap(pt)
			return
		}

		// Was connecting to a pin before?
		if c.potentialPin != nil {
			c.removePin(c.potentialPin)
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

	if c.subsumeInto != nil {
		c.view.changeSelection(c.subsumeInto)
		c.subsumeInto.commit()
	}
	if len(c.Pins) < 2 { // includes subsumption case
		c.view.changeSelection(c.graph)
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
	if p.channel != c {
		log.Print("c.removePin(p) but p.channel != c")
		return
	}
	if p == c.potentialPin {
		c.potentialPin = nil
	}
	p.unselected()
	p.channel = nil
	c.Pins[p].Remove()
	delete(c.Pins, p)
	c.cc.Detach(p.pc)
}

func (c *Channel) pinWasDeleted(p *Pin) {
	if p.channel != c {
		log.Print("c.pinWasDeleted(p) but p.channel != c")
		return
	}
	p.channel = nil
	c.Pins[p].Remove()
	delete(c.Pins, p)
	if len(c.Pins) < 2 {
		c.deleteView()
		return
	}
	c.layout(nil)
}

func (c *Channel) hasPin(p *Pin) bool {
	_, found := c.Pins[p]
	return found
}

func (c *Channel) subsume(ch *Channel) {
	ch.subsumeInto = c
	ch.presubsumption = make(map[*Pin]struct{})
	for p := range ch.Pins {
		ch.presubsumption[p] = struct{}{}
		ch.removePin(p)
		c.addPin(p)
	}
	c.layout(nil)
	ch.layout(nil)
	c.view.changeSelection(c)
}

func (c *Channel) unsubsume() {
	for p := range c.presubsumption {
		c.subsumeInto.removePin(p)
		c.addPin(p)
	}
	c.subsumeInto.layout(nil)
	c.layout(nil)
	c.view.changeSelection(c)
	c.subsumeInto = nil
	c.presubsumption = nil
}

func (c *Channel) gainFocus() {
	c.cc.GainFocus()
	c.Group.ClassList().Add("selected")
	for p := range c.Pins {
		p.selected()
	}
}

func (c *Channel) loseFocus() {
	go c.reallyCommit()
	c.Group.ClassList().Remove("selected")
	for p := range c.Pins {
		p.unselected()
	}
}

func (c *Channel) delete() {
	log.Print("TODO(josh): implement Channel.delete")
}

func (c *Channel) reallyDelete() {
	if err := c.cc.Delete(context.TODO()); err != nil {
		c.errors.setError("Couldn't delete channel: " + err.Error())
		return
	}

	c.deleteView()
}

func (c *Channel) deleteView() {
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
