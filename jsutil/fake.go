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

// MethodFunc is what a typical JS method looks like.
type MethodFunc func(...interface{}) *js.Object

// FakeObject implements a fake *js.Object.
type FakeObject struct {
	Properties map[string]*js.Object
	Methods    map[string]MethodFunc
}

// MakeFakeObject makes a FakeObject.
func MakeFakeObject() *FakeObject {
	return &FakeObject{
		Properties: make(map[string]*js.Object),
		Methods:    make(map[string]MethodFunc),
	}
}

// Get gets a property value.
func (o *FakeObject) Get(key string) *js.Object {
	return o.Properties[key]
}

// Set sets a property value.
func (o *FakeObject) Set(key string, value interface{}) {
	o.Properties[key] = js.MakeWrapper(value)
}

// Call calls a method.
func (o *FakeObject) Call(method string, args ...interface{}) *js.Object {
	return o.Methods[method](args...)
}

// FakeElement implements a virtual DOM element.
type FakeElement struct {
	FakeObject
	Class          string
	NamespaceURI   string
	Attributes     map[string]interface{}
	Children       []*FakeElement
	EventListeners map[string][]func(*js.Object)
}

// MakeFakeElement makes a fake element.
func MakeFakeElement(class, nsuri string) *FakeElement {
	return &FakeElement{
		FakeObject:     *MakeFakeObject(),
		Class:          class,
		NamespaceURI:   nsuri,
		Attributes:     make(map[string]interface{}),
		EventListeners: make(map[string][]func(*js.Object)),
	}
}

// ID returns e.Get("id").String() (so set the embedded FakeObject's property).
func (e *FakeElement) ID() string {
	return e.Get("id").String()
}

// SetAttribute sets an attribute.
func (e *FakeElement) SetAttribute(attr string, value interface{}) Element {
	e.Attributes[attr] = value
	return e
}

// RemoveAttribute removes an attribute.
func (e *FakeElement) RemoveAttribute(attr string) Element {
	delete(e.Attributes, attr)
	return e
}

// AddChildren adds child elements (*FakeElement only).
func (e *FakeElement) AddChildren(children ...Element) Element {
	for _, c := range children {
		if d, ok := c.(*FakeElement); ok {
			e.Children = append(e.Children, d)
		}
	}
	return e
}

// RemoveChildren removes child elements. It does it with the straightforward, naïve
// O(len(e.Children) * len(children)) method.
func (e *FakeElement) RemoveChildren(children ...Element) Element {
	if len(children) == 0 {
		return e
	}
	rem := make([]*FakeElement, 0, len(e.Children))
outer:
	for _, c := range e.Children {
		for _, x := range children {
			if c == x {
				continue outer
			}
		}
		rem = append(rem, c)
	}
	e.Children = rem
	return e
}

// AddEventListener adds an event listener.
func (e *FakeElement) AddEventListener(event string, handler func(*js.Object)) Element {
	e.EventListeners[event] = append(e.EventListeners[event], handler)
	return e
}

// Show removes the display attribute.
func (e *FakeElement) Show() Element {
	return e.RemoveAttribute("display")
}

// Hide sets the display attribute to "none".
func (e *FakeElement) Hide() Element {
	return e.SetAttribute("display", "none")
}

// FakeDocument implements a fake Document.
type FakeDocument struct {
	FakeElement
}

// MakeFakeDocument makes a fake document.
func MakeFakeDocument() *FakeDocument {
	return &FakeDocument{
		FakeElement: *MakeFakeElement("document", XHTMLNamespaceURI),
	}
}

// ElementByID searches the fake document for a matching element.
func (d *FakeDocument) ElementByID(id string) Element {
	stack := []*FakeElement{&d.FakeElement}
	for len(stack) > 0 {
		e := stack[0]
		if e.ID() == id {
			return e
		}
		stack = append(stack[1:], e.Children...)
	}
	return nil
}

// MakeTextNode makes something that looks like a text node.
func (d *FakeDocument) MakeTextNode(text string) Element {
	e := MakeFakeElement("text", "")
	e.Set("wholeText", text)
	return e
}

// MakeSVGElement makes an SVG element
func (d *FakeDocument) MakeSVGElement(class string) Element {
	return MakeFakeElement(class, SVGNamespaceURI)
}
