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

package jsutil

import "github.com/gopherjs/gopherjs/js"

// Element represents a DOM element. It embeds *js.Object.
type Element struct {
	*js.Object
}

// SetAttribute calls the JS setAttribute method, returning e for chaining.
func (e *Element) SetAttribute(attr string, value interface{}) *Element {
	e.Call("setAttribute", attr, value)
	return e
}

// AddChildren calls the JS method appendChild for each element, returning e for chaining.
func (e *Element) AddChildren(children ...*Element) *Element {
	for _, c := range children {
		e.Call("appendChild", c)
	}
	return e
}

// AddEventListener calls the JS method addEventListener, returning e for chaining.
func (e *Element) AddEventListener(event string, handler func(*js.Object)) *Element {
	e.Call("addEventListener", event, handler)
	return e
}

// Show sets the display attribute to the empty string.
func (e *Element) Show() *Element {
	e.Call("setAttribute", "display", "")
	return e
}

// Hide sets the display attribute to "none".
func (e *Element) Hide() *Element {
	e.Call("setAttribute", "display", "none")
	return e
}
