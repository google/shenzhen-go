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

	"github.com/google/shenzhen-go/dev/dom"
)

// Group includes handy methods for using SVG groups.
type Group struct {
	dom.Element
}

// NewGroup creates an SVG group in the given document.
func NewGroup(doc dom.Document) Group {
	return Group{doc.MakeSVGElement("g")}
}

// MoveTo moves the group to have the topleft corner at x, y.
func (g Group) MoveTo(x, y float64) Group {
	g.SetAttribute("transform", fmt.Sprintf("translate(%f, %f)", x, y))
	return g
}

// AddTo adds the group to the given parent element.
func (g Group) AddTo(parent dom.Element) Group {
	parent.AddChildren(g)
	return g
}

// Remove removes the group from the parent.
func (g Group) Remove() {
	g.Parent().RemoveChildren(g)
}
