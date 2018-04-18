// Copyright 2018 Google Inc.
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

import "github.com/google/shenzhen-go/dev/dom"

// Route is the visual connection between a pin and a channel.
type Route struct {
	line dom.Element

	ch *Channel
	p  *Pin
}

// NewRoute creates a route connecting a channel and a pin, and adds it
// as a child of the channel's group.
func NewRoute(doc dom.Document, ch *Channel, p *Pin) *Route {
	r := &Route{
		line: doc.MakeSVGElement("line").
			SetAttribute("stroke", normalColour).
			SetAttribute("stroke-width", lineWidth),
		ch: ch,
		p:  p,
	}
	r.Reroute()
	ch.Group.AddChildren(r.line)
	return r
}

// Remove removes the route.
func (r *Route) Remove() {
	r.ch.Group.RemoveChildren(r.line)
}

// Reroute repositions the route. Call after moving either the channel or the pin.
func (r *Route) Reroute() {
	r.line.
		SetAttribute("x1", r.ch.tx).
		SetAttribute("y1", r.ch.ty).
		SetAttribute("x2", r.p.x).
		SetAttribute("y2", r.p.y)
}

// Show shows the route.
func (r *Route) Show() { r.line.Show() }

// Hide hides the route.
func (r *Route) Hide() { r.line.Hide() }
