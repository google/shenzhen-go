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

//+build js

package dom

import "syscall/js"

// Well-known Namespace URIs.
const (
	SVGNamespaceURI   = "http://www.w3.org/2000/svg"
	XHTMLNamespaceURI = "http://www.w3.org/1999/xhtml"
)

// Document helps with some things the JS document global can do.
type Document struct {
	Element
}

// CurrentDocument returns the global document object or nil if it does not exist.
func CurrentDocument() Document {
	return Document{Element: Element{Value: js.Global().Get("document")}}
}

// ElementByID fetches an element by its id.
func (d Document) ElementByID(id string) Element {
	return Element{d.Call("getElementById", id)}
}

// MakeTextNode creates a new text node.
func (d Document) MakeTextNode(text string) Element {
	return Element{d.Call("createTextNode", text)}
}

// MakeSVGElement creates a new element in the SVG namespace.
func (d Document) MakeSVGElement(n string) Element {
	return Element{d.Call("createElementNS", SVGNamespaceURI, n)}
}
