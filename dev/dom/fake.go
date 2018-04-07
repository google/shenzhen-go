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

import (
	"math/rand"
	"reflect"
)

// MethodFunc is what a typical method looks like... sort of.
type MethodFunc func(...interface{}) interface{}

// FakeObject implements a fake Object that sort of works like *js.Object.
type FakeObject struct {
	Value      interface{}
	Properties map[string]interface{}
	Methods    map[string]MethodFunc
}

// MakeFakeObject makes a FakeObject.
func MakeFakeObject(value interface{}) *FakeObject {
	return &FakeObject{
		Value:      value,
		Properties: make(map[string]interface{}),
		Methods:    make(map[string]MethodFunc),
	}
}

// Get gets a "property value". In the case of fakes, this is a pointer that gets used as
// a key in ObjectsToValues.
func (o *FakeObject) Get(key string) Object {
	return MakeFakeObject(o.Properties[key])
}

// Set sets a property value.
func (o *FakeObject) Set(key string, value interface{}) { o.Properties[key] = value }

// Delete deletes the property with the given key.
func (o *FakeObject) Delete(key string) { delete(o.Properties, key) }

// Length returns the length of o.Value.
func (o *FakeObject) Length() int { return reflect.ValueOf(o.Value).Len() }

// Index returns the element at index i of o.Value.
func (o *FakeObject) Index(i int) Object {
	return MakeFakeObject(reflect.ValueOf(o.Value).Index(i).Interface())
}

// SetIndex sets the value at index i of o.Value to value.
func (o *FakeObject) SetIndex(i int, value interface{}) {
	reflect.ValueOf(o.Value).Index(i).Set(reflect.ValueOf(value))
}

// Call calls a method. In the case of fakes, this returns a pointer that gets used as
// a key in ObjectsToValues.
func (o *FakeObject) Call(method string, args ...interface{}) Object {
	return MakeFakeObject(o.Methods[method](args...))
}

// Invoke calls the function in o.Value.
func (o *FakeObject) Invoke(args ...interface{}) Object {
	a2 := make([]reflect.Value, len(args))
	for i, a := range args {
		a2[i] = reflect.ValueOf(a)
	}
	return MakeFakeObject(reflect.ValueOf(o.Value).CallSlice(a2))
}

// New calls the function in o.Value.
func (o *FakeObject) New(args ...interface{}) Object { return o.Invoke(args) }

// Bool returns o.Value asserted as a bool.
func (o *FakeObject) Bool() bool { return o.Value.(bool) }

// String returns o.Value asserted as a string.
func (o *FakeObject) String() string { return o.Value.(string) }

// Int returns o.Value asserted as an int.
func (o *FakeObject) Int() int { return o.Value.(int) }

// Int64 returns o.Value asserted as an int64.
func (o *FakeObject) Int64() int64 { return o.Value.(int64) }

// Uint64 returns o.Value asserted as a uint64.
func (o *FakeObject) Uint64() uint64 { return o.Value.(uint64) }

// Float returns o.Value asserted as a float64.
func (o *FakeObject) Float() float64 { return o.Value.(float64) }

// Interface returns o.Value.
func (o *FakeObject) Interface() interface{} { return o.Value }

// Unsafe returns o.Value asserted as a uintptr.
func (o *FakeObject) Unsafe() uintptr { return o.Value.(uintptr) }

// FakeElement implements a virtual DOM element.
type FakeElement struct {
	FakeObject
	Class          string
	NamespaceURI   string
	Attributes     map[string]interface{}
	Children       []*FakeElement
	EventListeners map[string][]func(Object)
	parent         *FakeElement
}

// MakeFakeElement makes a fake element.
func MakeFakeElement(class, nsuri string) *FakeElement {
	return &FakeElement{
		FakeObject:     *MakeFakeObject(nil),
		Class:          class,
		NamespaceURI:   nsuri,
		Attributes:     make(map[string]interface{}),
		EventListeners: make(map[string][]func(Object)),
	}
}

// ID returns e.Get("id").String() (so, set the embedded FakeObject's id property).
func (e *FakeElement) ID() string {
	return e.Get("id").String()
}

// SetAttribute sets an attribute.
func (e *FakeElement) SetAttribute(attr string, value interface{}) Element {
	e.Attributes[attr] = value
	// Also set the property, because property mapped attributes, but note this only
	// counts for some things.
	e.Properties[attr] = value
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
			d.parent = e
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
				c.parent = nil
				continue outer
			}
		}
		rem = append(rem, c)
	}
	e.Children = rem
	return e
}

// AddEventListener adds an event listener.
func (e *FakeElement) AddEventListener(event string, handler func(Object)) Element {
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

// Display sets the display attribute to whatever.
func (e *FakeElement) Display(style string) Element {
	return e.SetAttribute("display", style)
}

// Parent returns the parent element.
func (e *FakeElement) Parent() Element {
	return e.parent
}

// FakeDocument implements a fake Document.
type FakeDocument struct {
	FakeElement
}

// MakeFakeDocument makes a fake document.
func MakeFakeDocument() *FakeDocument {
	d := &FakeDocument{
		FakeElement: *MakeFakeElement("document", XHTMLNamespaceURI),
	}
	d.AddChildren(&FakeElement{
		Class: "body",
	})
	return d
}

// ElementByID searches the fake document for a matching element.
func (d *FakeDocument) ElementByID(id string) Element {
	q := []*FakeElement{&d.FakeElement}
	for len(q) > 0 {
		e := q[0]
		if e.ID() == id {
			return e
		}
		q = append(q[1:], e.Children...)
	}
	return nil
}

// MakeTextNode makes something that looks like a text node.
func (d *FakeDocument) MakeTextNode(text string) Element {
	e := MakeFakeElement("text", "")
	e.Set("wholeText", text)
	return e
}

// MakeSVGElement makes an SVG element.
func (d *FakeDocument) MakeSVGElement(class string) Element {
	e := MakeFakeElement(class, SVGNamespaceURI)
	switch class {
	case "text":
		e.Methods["getComputedTextLength"] = func(...interface{}) interface{} {
			return rand.Float64() * 200
		}
	}
	return e
}
