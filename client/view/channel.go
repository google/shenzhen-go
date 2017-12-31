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
	"regexp"
	"strconv"
	"strings"

	"github.com/google/shenzhen-go/jsutil"
	"github.com/google/shenzhen-go/model"
	pb "github.com/google/shenzhen-go/proto/js"
	"golang.org/x/net/context"
)

const anonChannelNamePrefix = "anonymousChannel"

var anonChannelNameRE = regexp.MustCompile(`^anonymousChannel\d+$`)

// Channel is the view's model of a channel.
type Channel struct {
	*model.Channel
	View *View

	// Cache of raw Pin objects which are connected.
	Pins map[*Pin]struct{}

	steiner jsutil.Element // symbol representing the channel itself, not used if channel is simple
	x, y    float64        // centre of steiner point, for snapping
	tx, ty  float64        // temporary centre of steiner point, for display
	l, c    jsutil.Element // for dragging to more pins
	p       *Pin           // considering attaching to this pin
}

func (v *View) newChannel(p, q *Pin) *Channel {
	c := &model.Channel{
		Type:      p.Type,
		Capacity:  0,
		Anonymous: true,
	}
	// Pick a unique name
	max := -1
	for _, ec := range v.Graph.Channels {
		if anonChannelNameRE.MatchString(ec.Name) {
			n, err := strconv.Atoi(strings.TrimPrefix(ec.Name, anonChannelNamePrefix))
			if err != nil {
				log.Printf("Couldn't convert digits into an int: %v", err)
				return nil
			}
			if n > max {
				max = n
			}
		}
	}
	c.Name = anonChannelNamePrefix + strconv.Itoa(max+1)
	go func() {
		req := &pb.CreateChannelRequest{
			Graph: v.Graph.FilePath,
			Name:  c.Name,
			Type:  c.Type,
			Cap:   uint64(c.Capacity),
			Anon:  c.Anonymous,
			Node1: p.node.Name,
			Pin1:  p.Name,
			Node2: q.node.Name,
			Pin2:  q.Name,
		}
		if _, err := v.Client.CreateChannel(context.Background(), req); err != nil {
			log.Printf("Couldn't CreateChannel: %v", err)
		}
	}()
	ch := &Channel{
		Channel: c,
		View:    v,
		Pins: map[*Pin]struct{}{
			p: {},
			q: {},
		},
	}
	ch.makeElements()
	return ch
}

func (c *Channel) makeElements() {
	c.steiner = c.View.Document.MakeSVGElement("circle").
		SetAttribute("r", pinRadius).
		AddEventListener("mousedown", c.dragStart)

	c.l = c.View.Document.MakeSVGElement("line").
		SetAttribute("stroke-width", lineWidth).
		Hide()

	c.c = c.View.Document.MakeSVGElement("circle").
		SetAttribute("r", pinRadius).
		SetAttribute("fill", "transparent").
		SetAttribute("stroke-width", lineWidth).
		Hide()

	c.View.Diagram.AddChildren(c.steiner, c.l, c.c)
}

func (c *Channel) unmakeElements() {
	c.View.Diagram.RemoveChildren(c.steiner, c.l, c.c)
	c.steiner, c.l, c.c = nil, nil, nil
}

// Pt implements Point.
func (c *Channel) Pt() (x, y float64) { return c.x, c.y }

func (c *Channel) commit() { c.x, c.y = c.tx, c.ty }

func (c *Channel) dragStart(e jsutil.Object) {
	c.View.Diagram.dragItem = c

	// TODO: make it so that if the current configuration is invalid
	// (e.g. all input pins / output pins) then use errorColour, and
	// delete the whole channel if dropped.

	c.steiner.Show()
	c.setColour(activeColour)

	x, y := c.View.Diagram.cursorPos(e)
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

func (c *Channel) drag(e jsutil.Object) {
	x, y := c.View.Diagram.cursorPos(e)
	c.steiner.Show()
	c.l.
		SetAttribute("x1", x).
		SetAttribute("y1", y)
	c.c.
		SetAttribute("cx", x).
		SetAttribute("cy", y)
	d, q := c.View.Graph.nearestPoint(x, y)
	p, _ := q.(*Pin)

	if p != nil && p == c.p && d < snapQuad {
		return
	}

	if c.p != nil && (c.p != p || d >= snapQuad) {
		c.p.disconnect()
		c.p.circ.SetAttribute("fill", normalColour)
		c.p.l.Hide()
		c.p = nil
	}

	noSnap := func() {
		c.c.Show()
		c.l.Show()
		c.reposition(ephemeral{x, y})
	}

	if d >= snapQuad || q == c || (p != nil && p.ch == c) {
		c.View.Diagram.clearError()
		noSnap()
		c.setColour(activeColour)
		return
	}

	if p == nil || p.ch != nil {
		c.View.Diagram.setError("Can't connect different channels together (use another goroutine)", x, y)
		noSnap()
		c.setColour(errorColour)
		return
	}

	if err := p.connectTo(c); err != nil {
		c.View.Diagram.setError("Can't connect: "+err.Error(), x, y)
		noSnap()
		c.setColour(errorColour)
		return
	}

	// Let's snap!
	c.View.Diagram.clearError()
	c.p = p
	p.l.Show()
	c.setColour(activeColour)
	c.l.Hide()
	c.c.Hide()
}

func (c *Channel) drop(e jsutil.Object) {
	c.View.Diagram.clearError()
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

func (c *Channel) gainFocus(e jsutil.Object) {
	log.Print("TODO(josh): implement Channel.gainFocus")
}

func (c *Channel) loseFocus(e jsutil.Object) {
	log.Print("TODO(josh): implement Channel.loseFocus")
}

func (c *Channel) save(e jsutil.Object) {
	log.Print("TODO(josh): implement Channel.save")
}

func (c *Channel) delete(e jsutil.Object) {
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
		for t := range c.Pins {
			t.c.Hide()
			t.l.Hide()
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
	for t := range c.Pins {
		t.l.
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
		t.circ.SetAttribute("fill", col)
		t.l.SetAttribute("stroke", col)
	}
}
