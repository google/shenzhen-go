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

package main

import (
	"github.com/gopherjs/gopherjs/js"
)

const (
	activeColour = "#09f"
	normalColour = "#000"
	errorColour  = "#f06"

	pinRadius = 5
	lineWidth = 2
	snapQuad  = 144
)

var (
	document   = js.Global.Get("document")
	diagramSVG = document.Call("getElementById", "diagram")
	svgNS      = diagramSVG.Get("namespaceURI")

	apiEndpoint = js.Global.Get("apiEndpoint").String()

	dragItem interface{}

	graph = &Graph{
		Nodes: map[string]*Node{
			"Hello, yes": {
				Name: "Hello, yes",
				Inputs: []*Pin{
					{Name: "foo1", Type: "int", input: true},
					{Name: "foo2", Type: "int", input: true},
					{Name: "foo3", Type: "string", input: true},
				},
				Outputs: []*Pin{
					{Name: "bar", Type: "string"},
					{Name: "baz", Type: "string"},
				},
				X: 100,
				Y: 100,
			},
			"this is dog": {
				Name: "this is dog",
				Inputs: []*Pin{
					{Name: "baz0", Type: "string", input: true},
					{Name: "baz1", Type: "string", input: true},
					{Name: "baz2", Type: "string", input: true},
					{Name: "baz3", Type: "string", input: true},
				},
				Outputs: []*Pin{
					{Name: "qux", Type: "float64"},
				},
				X: 100,
				Y: 200,
			},
		},
		Channels: map[*Channel]struct{}{},
	}
)

func makeSVGElement(n string) *js.Object { return document.Call("createElementNS", svgNS, n) }
func cursorPos(e *js.Object) (x, y float64) {
	bcr := diagramSVG.Call("getBoundingClientRect")
	x = e.Get("clientX").Float() - bcr.Get("left").Float()
	y = e.Get("clientY").Float() - bcr.Get("top").Float()
	return
}

type Point interface {
	Pt() (x, y float64)
}

type ephemeral struct{ x, y float64 }

func (e ephemeral) Pt() (x, y float64) { return e.x, e.y }

func mouseMove(e *js.Object) {
	if dragItem == nil {
		return
	}
	switch t := dragItem.(type) {
	case *Node:
		t.moveTo(e.Get("clientX").Float()-t.relX, e.Get("clientY").Float()-t.relY)
	case *Channel:
		t.dragTo(e)
	case *Pin:
		t.dragTo(e)
	}
}

func mouseUp(e *js.Object) {
	if dragItem == nil {
		return
	}
	mouseMove(e)

	switch t := dragItem.(type) {
	case *Node:
		// Nothing
	case *Channel:
		t.drop(e)
	case *Pin:
		t.drop(e)
	}
	dragItem = nil
}

func main() {
	if apiEndpoint != "" {
		loadGraph()
	}

	for _, n := range graph.Nodes {
		n.makeNodeElement()
	}

	diagramSVG.Call("addEventListener", "mousemove", mouseMove)
	diagramSVG.Call("addEventListener", "mouseup", mouseUp)
}
