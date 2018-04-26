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

// Route is the visual connection between two points
type Route struct {
	line dom.Element

	src, dst Pointer
	parent   Group
}

// NewRoute creates a route connecting a channel and a pin, and adds it
// as a child of the channel's group.
func NewRoute(doc dom.Document, parent Group, src, dst Pointer) *Route {
	r := &Route{
		line: doc.MakeSVGElement("line").
			SetAttribute("stroke", normalColour).
			SetAttribute("stroke-width", lineWidth),
		src: src,
		dst: dst,
	}
	r.Reroute()
	parent.AddChildren(r.line)
	return r
}

// Remove removes the route.
func (r *Route) Remove() {
	if r == nil || r.line == nil {
		return
	}
	p := r.line.Parent()
	if p == nil {
		return
	}
	p.RemoveChildren(r.line)
}

// Reroute repositions the route. Call after moving either the channel or the pin.
func (r *Route) Reroute() {
	a, b := r.src.Pt(), r.dst.Pt()
	r.line.
		SetAttribute("x1", real(a)).
		SetAttribute("y1", imag(a)).
		SetAttribute("x2", real(b)).
		SetAttribute("y2", imag(b))
}

// SetStroke sets the stroke colour.
func (r *Route) SetStroke(colour string) { r.line.SetAttribute("stroke", colour) }

// Show shows the route.
func (r *Route) Show() { r.line.Show() }

// Hide hides the route.
func (r *Route) Hide() { r.line.Hide() }
