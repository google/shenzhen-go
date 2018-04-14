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
	"errors"

	"github.com/google/shenzhen-go/dev/dom"
	pb "github.com/google/shenzhen-go/dev/proto/js"
	"golang.org/x/net/context"
)

const (
	nametagRectStyle = "fill: #efe; stroke: #353; stroke-width:1"
	nametagTextStyle = "font-family:Go; font-size:16; user-select:none; pointer-events:none"
)

// Pin represents a node pin visually, and has enough information to know
// if it is validly connected.
type Pin struct {
	Group
	Shape   dom.Element // my main visual representation
	Nametag *TextBox    // Hello, my name is ...
	x, y    float64     // computed, not relative to node

	// TODO: consult a controller?
	Name, Type string
	input      bool // am I an input?

	node *Node    // owner.
	ch   *Channel // attached to this channel, is often nil
}

// Save time by checking whether a potential connection can succeeds.
func (p *Pin) checkConnectionTo(q Point) error {
	switch q := q.(type) {
	case *Pin:
		if q.Type != p.Type {
			return errors.New("mismatching types [" + p.Type + " != " + q.Type + "]")
		}
		if q.ch != nil {
			return p.checkConnectionTo(q.ch)
		}

		// Prevent mistakes by ensuring that there is at least one input
		// and one output per channel, and they connect separate goroutines.
		if p.input == q.input {
			return errors.New("both pins have the same direction")
		}
		if p.node == q.node {
			return errors.New("both pins are on the same goroutine")
		}

	case *Channel:
		if q.cc.Channel().Type != p.Type {
			return errors.New("mismatching types [" + p.Type + " != " + q.cc.Channel().Type + "]")
		}
		same := true
		for r := range q.Pins {
			if r.input != p.input {
				same = false
				break
			}
		}
		if same {
			return errors.New("must connect at least one input and one output")
		}
	}
	return nil
}

func (p *Pin) reallyConnect() {
	// Attach to the existing channel
	if _, err := p.node.view.client.ConnectPin(context.Background(), &pb.ConnectPinRequest{
		Graph:   p.node.view.graph.gc.Graph().FilePath,
		Node:    p.node.nc.Node().Name,
		Pin:     p.Name,
		Channel: p.ch.cc.Channel().Name,
	}); err != nil {
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
	if _, err := p.node.view.client.DisconnectPin(context.Background(), &pb.DisconnectPinRequest{
		Graph: p.node.view.graph.gc.Graph().FilePath,
		Node:  p.node.nc.Node().Name,
		Pin:   p.Name,
	}); err != nil {
		p.node.view.setError("Couldn't disconnect: " + err.Error())
	}
}

// MoveTo moves the pin (relatively).
func (p *Pin) MoveTo(rx, ry float64) {
	p.Shape.SetAttribute("cx", rx).SetAttribute("cy", ry)
	p.x, p.y = rx+p.node.nc.Node().X, ry+p.node.nc.Node().Y
}

// Pt returns the diagram coordinate of the pin, for nearest-neighbor purposes.
func (p *Pin) Pt() (x, y float64) { return p.x, p.y }

func (p *Pin) String() string { return p.node.nc.Node().Name + "." + p.Name }

func (p *Pin) dragStart(e dom.Object) {
	// If a channel is attached, detach and drag from that instead.
	// If not, create a new channela and attach it.
	if p.ch != nil {
		p.disconnect()
		p.ch.dragStart(e)
		return
	}
	p.ch = &Channel{
		// TODO
		Pins: map[*Pin]*Route{
			p: nil,
		},
	}
	p.ch.dragStart(e)
}

func (p *Pin) mouseEnter(dom.Object) {
	x, y := p.x-p.node.nc.Node().X, p.y-p.node.nc.Node().Y
	if p.input {
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
	p.Group.AddChildren(p.Shape)

	// Nametag textbox.
	p.Nametag = &TextBox{Margin: 20, TextOffsetY: 5}
	p.Nametag.
		MakeElements(doc, p.Group).
		SetHeight(30).
		SetTextStyle(nametagTextStyle).
		SetRectStyle(nametagRectStyle).
		SetText(p.Name + " (" + p.Type + ")")
	p.Nametag.RecomputeWidth()
	p.Nametag.Hide()

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
