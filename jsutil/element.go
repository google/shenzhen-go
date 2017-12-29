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

// Object is some stuff JS objects can do.
type Object interface {
	Get(string) *js.Object
	Set(string, interface{})
	Call(string, ...interface{}) *js.Object
}

// Element represents a DOM element.
type Element interface {
	Object
	SetAttribute(string, interface{}) Element
	AddChildren(...Element) Element
	RemoveChildren(...Element) Element
	AddEventListener(string, func(*js.Object)) Element
	Show() Element
	Hide() Element
}

type element struct {
	*js.Object
}

// Wrap turns a *js.Object into an Element.
func Wrap(o *js.Object) Element { return &element{o} }

// SetAttribute calls the JS setAttribute method, returning e for chaining.
func (e *element) SetAttribute(attr string, value interface{}) Element {
	e.Call("setAttribute", attr, value)
	return e
}

// AddChildren calls the JS method appendChild for each element, returning e for chaining.
func (e *element) AddChildren(children ...Element) Element {
	for _, c := range children {
		e.Call("appendChild", c)
	}
	return e
}

// RemoveChildren calls the JS method removeChild for each element, returning e for chaining.
func (e *element) RemoveChildren(children ...Element) Element {
	for _, c := range children {
		e.Call("removeChild", c)
	}
	return e
}

// AddEventListener calls the JS method addEventListener, returning e for chaining.
func (e *element) AddEventListener(event string, handler func(*js.Object)) Element {
	e.Call("addEventListener", event, handler)
	return e
}

// Show removes the display attribute.
func (e *element) Show() Element {
	e.Call("removeAttribute", "display")
	return e
}

// Hide sets the display attribute to "none".
func (e *element) Hide() Element {
	e.Call("setAttribute", "display", "none")
	return e
}
