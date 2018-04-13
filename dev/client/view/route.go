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
	dom.Element
}

// MakeElements makes SVG elements for the route and attaches them to the DOM.
func (r *Route) MakeElements(doc dom.Document, ch *Channel) {
	r.Element = doc.MakeSVGElement("line").
		SetAttribute("stroke", normalColour).
		SetAttribute("x1", ch.tx).
		SetAttribute("y1", ch.ty)
	ch.Group.AddChildren(r.Element)
}
