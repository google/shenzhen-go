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
	"sort"
	"strings"
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
	if o, ok := value.(*FakeObject); ok {
		return o
	}
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

// Float returns o.Value asserted as a float64.
func (o *FakeObject) Float() float64 { return o.Value.(float64) }

// FakeClassList implements a virtual classList (DOMTokenList).
// Unlike a real DOMTokenList, it doesn't preserve order.
type FakeClassList map[string]struct{}

// Add adds a class to the classlist.
func (c FakeClassList) Add(classes ...string) {
	for _, cl := range classes {
		c[cl] = struct{}{}
	}
}

// Remove removes a class from the classlist.
func (c FakeClassList) Remove(classes ...string) {
	for _, cl := range classes {
		delete(c, cl)
	}
}

// Toggle adds if the class is not present, and removes if it is.
func (c FakeClassList) Toggle(class string) {
	if c.Contains(class) {
		c.Remove(class)
	} else {
		c.Add(class)
	}
}

// Contains tests if the class is present.
func (c FakeClassList) Contains(class string) bool {
	_, found := c[class]
	return found
}

// Replace swaps an old class for a new one.
func (c FakeClassList) Replace(oldClass, newClass string) {
	c.Remove(oldClass)
	c.Add(newClass)
}

func (c FakeClassList) String() string {
	cls := make([]string, 0, len(c))
	for cl := range c {
		cls = append(cls, cl)
	}
	sort.Strings(cls)
	return strings.Join(cls, " ")
}

// FakeCallback is a fake Callback.
type FakeCallback struct {
	Fn func([]Object)
}

// Release does nothing (it does in real JS land).
func (FakeCallback) Release() {}

// FakeEventCallback is another fake Callback.
type FakeEventCallback struct {
	Fn func(Object)
}

// Release does nothing (it does in real JS land).
func (FakeEventCallback) Release() {}

// FakeElement implements a virtual DOM element.
type FakeElement struct {
	FakeObject
	Class          string
	NamespaceURI   string
	Attributes     map[string]interface{}
	Children       []*FakeElement
	EventListeners map[string][]Callback
	Classes        FakeClassList
	parent         *FakeElement
}

// MakeFakeElement makes a fake element.
func MakeFakeElement(class, nsuri string) *FakeElement {
	return &FakeElement{
		FakeObject:     *MakeFakeObject(nil),
		Class:          class,
		NamespaceURI:   nsuri,
		Attributes:     make(map[string]interface{}),
		EventListeners: make(map[string][]Callback),
		Classes:        make(FakeClassList),
	}
}

// ID returns e.Get("id").String() (so, set the embedded FakeObject's id property).
func (e *FakeElement) ID() string {
	return e.Get("id").String()
}

// GetAttribute gets an attribute value.
func (e *FakeElement) GetAttribute(attr string) Object {
	return MakeFakeObject(e.Attributes[attr])
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
func (e *FakeElement) AddEventListener(event string, cb Callback) Element {
	e.EventListeners[event] = append(e.EventListeners[event], cb)
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

// ClassList returns the list of classes.
func (e *FakeElement) ClassList() ClassList {
	return e.Classes
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
		e.Methods["getBBox"] = func(...interface{}) interface{} {
			o := MakeFakeObject(nil)
			o.Set("x", 150.0)
			o.Set("y", 160.0)
			o.Set("height", 40.0)
			o.Set("width", 130.0)
			return o
		}
	}
	return e
}
