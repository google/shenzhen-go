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

import (
	"fmt"
	"math/cmplx"

	"github.com/google/shenzhen-go/dev/dom"
)

const arrowSize = 5

// Route is the visual connection between two points
type Route struct {
	Group
	line  dom.Element
	arrow dom.Element

	src, dst Pointer
}

// NewRoute creates a route connecting a channel and a pin, and adds it
// as a child of the channel's group.
func NewRoute(doc dom.Document, parent Group, src, dst Pointer) *Route {
	r := &Route{
		Group: NewGroup(doc, parent),
		line:  doc.MakeSVGElement("line"),
		arrow: doc.MakeSVGElement("path"),
		src:   src,
		dst:   dst,
	}
	r.Group.AddChildren(r.line, r.arrow)
	r.line.ClassList().Add("route")
	r.arrow.ClassList().Add("route", "arrow")
	r.Reroute()
	return r
}

// Reroute repositions the route. Call after moving either the channel or the pin.
func (r *Route) Reroute() {
	a, b := r.src.Pt(), r.dst.Pt()
	r.line.
		SetAttribute("x1", real(a)).
		SetAttribute("y1", imag(a)).
		SetAttribute("x2", real(b)).
		SetAttribute("y2", imag(b))

	// Manually-managed arrow symbol.
	d := complex128(b - a)
	md := cmplx.Abs(d)
	// Don't show the symbol if it's too crowded.
	if md < 3*arrowSize {
		r.arrow.Hide()
		return
	}
	r.arrow.Show()
	c := complex128(a+b) / 2.0
	scale := complex(arrowSize/md, 0)
	d = scale * d // Scaled unit vector in the direction a -> b
	p := d * 1i   // Perpendicular to d.
	c0, c1, c2 := c+d, c-d+p, c-d-p
	r.arrow.SetAttribute("d", fmt.Sprintf(
		"M %f %f L %f %f L %f %f z",
		real(c0), imag(c0),
		real(c1), imag(c1),
		real(c2), imag(c2)))
}
