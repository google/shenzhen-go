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

// Well-known Namespace URIs.
const (
	SVGNamespaceURI   = "http://www.w3.org/2000/svg"
	XHTMLNamespaceURI = "http://www.w3.org/1999/xhtml"
)

// Document describes some things the JS document global can do.
type Document interface {
	Element

	ElementByID(string) Element
	MakeTextNode(string) Element
	MakeSVGElement(string) Element
}

type document struct {
	Element
}

// CurrentDocument returns the global document object or nil if it does not exist.
func CurrentDocument() Document {
	d := js.Global.Get("document")
	if d == nil {
		return nil
	}
	return document{Element: WrapElement(WrapObject(d))}
}

func (d document) ElementByID(id string) Element {
	return WrapElement(d.Call("getElementById", id))
}

func (d document) MakeTextNode(text string) Element {
	return WrapElement(d.Call("createTextNode", text))
}

func (d document) MakeSVGElement(n string) Element {
	return WrapElement(d.Call("createElementNS", SVGNamespaceURI, n))
}
