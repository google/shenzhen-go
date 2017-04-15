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
	"log"

	"github.com/google/shenzhen-go/api"
	"github.com/gopherjs/gopherjs/js"
)

const (
	activeColour = "#09f"
	normalColour = "#000"
	errorColour  = "#f06"

	errRectStyle = "fill: #fee; fill-opacity: 0.5; stroke: #533; stroke-width:1"
	errTextStyle = "font-family:Go; font-size:16; user-select:none; pointer-events:none"

	pinRadius = 5
	lineWidth = 2
	snapQuad  = 144
)

var (
	document = js.Global.Get("document")

	graphPath = js.Global.Get("graphPath").String()

	client api.Interface
)

type diagram struct {
	*js.Object

	dragItem draggable
	errLabel *textBox
	graph    *Graph
}

func (d *diagram) makeSVGElement(n string) *js.Object {
	return document.Call("createElementNS", d.Get("namespaceURI"), n)
}

func (d *diagram) cursorPos(e *js.Object) (x, y float64) {
	bcr := d.Call("getBoundingClientRect")
	x = e.Get("clientX").Float() - bcr.Get("left").Float()
	y = e.Get("clientY").Float() - bcr.Get("top").Float()
	return
}

// Point is anything that has a position on the canvas.
type Point interface {
	Pt() (x, y float64)
}

type ephemeral struct{ x, y float64 }

func (e ephemeral) Pt() (x, y float64) { return e.x, e.y }

type draggable interface {
	drag(*js.Object)
	drop(*js.Object)
}

func (d *diagram) mouseMove(e *js.Object) {
	if d.dragItem == nil {
		return
	}
	d.dragItem.drag(e)
}

func (d *diagram) mouseUp(e *js.Object) {
	if d.dragItem == nil {
		return
	}
	d.dragItem.drag(e)
	d.dragItem.drop(e)
	d.dragItem = nil
}

func (d *diagram) setError(err string, x, y float64) {
	if err == "" {
		d.clearError()
		return
	}
	d.Call("appendChild", d.errLabel.group) // Bring to front
	d.errLabel.moveTo(x+4, y-36)
	d.errLabel.setText(err)
	d.errLabel.show()
}

func (d *diagram) clearError() {
	d.errLabel.hide()
}

func main() {
	apiURL := js.Global.Get("apiURL").String()

	client = api.NewClient(apiURL)

	d := &diagram{
		Object: document.Call("getElementById", "diagram"),
	}
	if d.Object == nil {
		log.Fatalf("Couldn't find diagram element")
	}
	d.errLabel = newTextBox(d, "", errTextStyle, errRectStyle, 0, 0, 0, 32)
	d.errLabel.hide()

	g, err := loadGraph(d)
	if err != nil {
		log.Fatalf("Couldn't load graph: %v", err)
	}
	d.graph = g

	d.Call("addEventListener", "mousemove", d.mouseMove)
	d.Call("addEventListener", "mouseup", d.mouseUp)

	sgp := document.Call("getElementById", "save-graph-properties")
	if sgp == nil {
		log.Fatalf("Couldn't find save-graph-properties element")
	}
	sgp.Call("addEventListener", "click", d.graph.saveProperties)
}
