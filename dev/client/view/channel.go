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

	cc    ChannelController
	view  *View
	graph *Graph

	// Cache of raw Pin objects which are connected.
	Pins map[*Pin]*Route

	created bool // create operation sent to server?

	steiner dom.Element // symbol representing the channel itself, not used if channel is simple
	x, y    float64     // centre of steiner point, for snapping
	tx, ty  float64     // temporary centre of steiner point, for display
	l, c    dom.Element // for dragging to more pins
	p       *Pin        // considering attaching to this pin
}

/*
func (v *View) createChannel(p *Pin) *Channel {
	cc, err := v.graph.gc.CreateChannel(p.node.nc.Name(), p.Name)
	if err != nil {
		// TODO: handle better
		panic(err)
	}
	ch := &Channel{
		cc:   cc,
		view: v,
		Pins: map[*Pin]*Route{
			p: {},
		},
	}
	ch.makeElements(v.doc, v.diagram)
	return ch
}
*/

func (c *Channel) reallyCreate() {
	if err := c.cc.Commit(context.TODO()); err != nil {
		c.view.setError("Couldn't create a channel: " + err.Error())
		return
	}
	c.created = true
}

func (c *Channel) makeElements(doc dom.Document, parent dom.Element) {
	c.Group = NewGroup(doc, parent)

	c.steiner = doc.MakeSVGElement("circle").
		SetAttribute("r", pinRadius).
		AddEventListener("mousedown", c.dragStart)

	c.l = doc.MakeSVGElement("line").
		SetAttribute("stroke-width", lineWidth).
		Hide()

	c.c = doc.MakeSVGElement("circle").
		SetAttribute("r", pinRadius).
		SetAttribute("fill", "transparent").
		SetAttribute("stroke-width", lineWidth).
		Hide()

	c.Group.AddChildren(c.steiner, c.l, c.c)
}

// Pt implements Point.
func (c *Channel) Pt() (x, y float64) { return c.x, c.y }

func (c *Channel) commit() {
	c.x, c.y = c.tx, c.ty
	if !c.created {
		go c.reallyCreate()
	}
}

func (c *Channel) dragStart(e dom.Object) {
	c.view.diagram.dragItem = c

	// TODO: make it so that if the current configuration is invalid
	// (e.g. all input pins / output pins) then use errorColour, and
	// delete the whole channel if dropped.

	c.steiner.Show()
	c.setColour(activeColour)

	x, y := c.view.diagram.cursorPos(e)
	c.reposition(ephemeral{x, y})
	c.l.
		SetAttribute("x1", x).
		SetAttribute("y1", y).
		SetAttribute("x2", c.tx).
		SetAttribute("y2", c.ty).
		Show()

	c.c.
		SetAttribute("cx", x).
		SetAttribute("cy", y).
		Show()
}

func (c *Channel) drag(e dom.Object) {
	x, y := c.view.diagram.cursorPos(e)
	c.steiner.Show()
	c.l.
		SetAttribute("x1", x).
		SetAttribute("y1", y)
	c.c.
		SetAttribute("cx", x).
		SetAttribute("cy", y)
	d, q := c.view.graph.nearestPoint(x, y)
	p, _ := q.(*Pin)

	if p != nil && p == c.p && d < snapQuad {
		return
	}

	if c.p != nil && (c.p != p || d >= snapQuad) {
		c.p.disconnect()
		c.p.Shape.SetAttribute("fill", normalColour)
		c.p = nil
	}

	noSnap := func() {
		c.c.Show()
		c.l.Show()
		c.reposition(ephemeral{x, y})
	}

	if d >= snapQuad || q == c || (p != nil && p.ch == c) {
		c.view.diagram.clearError()
		noSnap()
		c.setColour(activeColour)
		return
	}

	if p == nil || p.ch != nil {
		c.view.diagram.setError("Can't connect different channels together (use another goroutine)", x, y)
		noSnap()
		c.setColour(errorColour)
		return
	}

	if err := p.checkConnectionTo(c); err != nil {
		c.view.diagram.setError("Can't connect: "+err.Error(), x, y)
		noSnap()
		c.setColour(errorColour)
		return
	}

	// Let's snap!
	c.view.diagram.clearError()
	c.p = p
	c.setColour(activeColour)
	c.l.Hide()
	c.c.Hide()
}

func (c *Channel) drop(e dom.Object) {
	c.view.diagram.clearError()
	c.reposition(nil)
	c.commit()
	c.setColour(normalColour)
	if c.p != nil {
		c.p = nil
		return
	}
	c.c.Hide()
	c.l.Hide()
	if len(c.Pins) <= 2 {
		c.steiner.Hide()
	}
}

func (c *Channel) gainFocus(e dom.Object) {
	log.Print("TODO(josh): implement Channel.gainFocus")
}

func (c *Channel) loseFocus(e dom.Object) {
	log.Print("TODO(josh): implement Channel.loseFocus")
}

func (c *Channel) save(e dom.Object) {
	log.Print("TODO(josh): implement Channel.save")
}

func (c *Channel) delete(e dom.Object) {
	log.Print("TODO(josh): implement Channel.delete")
}

func (c *Channel) reposition(additional Point) {
	np := len(c.Pins)
	if additional != nil {
		np++
	}
	if np < 2 {
		// Not actually a channel anymore - hide.
		c.steiner.Hide()
		for _, r := range c.Pins {
			r.Element.Hide()
		}
		return
	}
	c.tx, c.ty = 0, 0
	if additional != nil {
		c.tx, c.ty = additional.Pt()
	}
	for t := range c.Pins {
		c.tx += t.x
		c.ty += t.y
	}
	n := float64(np)
	c.tx /= n
	c.ty /= n
	c.steiner.
		SetAttribute("cx", c.tx).
		SetAttribute("cy", c.ty)
	c.l.
		SetAttribute("x2", c.tx).
		SetAttribute("y2", c.ty)
	for _, r := range c.Pins {
		r.Element.
			SetAttribute("x2", c.tx).
			SetAttribute("y2", c.ty)
	}
	if np <= 2 {
		c.steiner.Hide()
	} else {
		c.steiner.Show()
	}
}

func (c *Channel) setColour(col string) {
	c.steiner.SetAttribute("fill", col)
	c.c.SetAttribute("stroke", col)
	c.l.SetAttribute("stroke", col)
	for t := range c.Pins {
		t.Shape.SetAttribute("fill", col)
		//t.l.SetAttribute("stroke", col)
	}
}
