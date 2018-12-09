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

package dom

import "syscall/js"

// Element represents a DOM element.
type Element struct {
	js.Value
}

// ID returns the element ID.
func (e Element) ID() string {
	return e.Get("id").String()
}

// GetAttribute gets the JS getAttribute method, returning the requested attribute.
func (e Element) GetAttribute(attr string) js.Value {
	return e.Call("getAttribute", attr)
}

// SetAttribute calls the JS setAttribute method, returning the element for chaining.
func (e Element) SetAttribute(attr string, value interface{}) Element {
	e.Call("setAttribute", attr, value)
	return e
}

// RemoveAttribute calls the JS removeAttribute method, returning the element for chaining.
func (e Element) RemoveAttribute(attr string) Element {
	e.Call("removeAttribute", attr)
	return e
}

// AddChildren calls the JS method appendChild for each element, returning the element for chaining.
func (e Element) AddChildren(children ...Element) Element {
	for _, c := range children {
		e.Call("appendChild", c.Value)
	}
	return e
}

// RemoveChildren calls the JS method removeChild for each element, returning the element for chaining.
func (e Element) RemoveChildren(children ...Element) Element {
	for _, c := range children {
		e.Call("removeChild", c.Value)
	}
	return e
}

// AddEventListener calls the JS method addEventListener, returning the element for chaining.
func (e Element) AddEventListener(event string, cb js.Callback) Element {
	e.Call("addEventListener", event, cb)
	return e
}

// Show sets the display attribute of the style to "", returning the element for chaining.
func (e Element) Show() Element { return e.Display("") }

// Hide sets the display attribute of the style to "none", returning the element for chaining.
func (e Element) Hide() Element { return e.Display("none") }

// Display sets the display attribute of the style to the given value, returning the element for chaining.
func (e Element) Display(style string) Element {
	e.Get("style").Set("display", style)
	return e
}

// Parent returns the parent element.
func (e Element) Parent() Element {
	return Element{Value: e.Get("parentElement")}
}

// ClassList returns the classList.
func (e Element) ClassList() ClassList {
	return ClassList{Value: e.Get("classList")}
}
