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

// Package view implements the client.
package view

import (
	"log"

	"github.com/google/shenzhen-go/dev/model"

	"github.com/google/shenzhen-go/dev/dom"
)

const (
	activeColour = "#09f"
	normalColour = "#000"
	errorColour  = "#f06"

	errRectStyle = "fill: #fee; stroke: #533; stroke-width:1"
	errTextStyle = "font-family:Go; font-size:16; user-select:none; pointer-events:none"

	pinRadius = 5
	lineWidth = 2
	snapQuad  = 144
)

// View caches the top-level objects for managing the UI.
type View struct {
	doc     dom.Document // Global document object
	diagram dom.Element  // The LHS panel
	graph   *Graph       // SVG elements in the LHS panel

	dragItem     draggable  // nil if nothing is being dragged
	selectedItem selectable // nil if nothing is selected
	errLabel     *TextBox
}

// Setup connects to elements in the DOM.
func Setup(doc dom.Document, gc GraphController) error {
	v := &View{
		doc:     doc,
		diagram: doc.ElementByID("diagram"),
	}

	v.graph = &Graph{
		gc:   gc,
		doc:  doc,
		view: v,
	}
	v.graph.makeElements(doc, v.diagram)
	v.graph.refresh()

	v.diagram.
		AddEventListener("mousedown", v.diagramMouseDown).
		AddEventListener("mousemove", v.diagramMouseMove).
		AddEventListener("mouseup", v.diagramMouseUp)

	doc.ElementByID("graph-save").
		AddEventListener("click", v.graph.save)
	doc.ElementByID("graph-properties-save").
		AddEventListener("click", v.graph.saveProperties)

	doc.ElementByID("node-save-link").
		AddEventListener("click", v.saveSelected)
	doc.ElementByID("node-delete-link").
		AddEventListener("click", v.deleteSelected)

	doc.ElementByID("node-metadata-link").
		AddEventListener("click", func(dom.Object) {
			v.selectedItem.(*Node).nc.ShowMetadataSubpanel()
		})

	for n, t := range model.PartTypes {
		doc.ElementByID("node-new-link:"+n).
			AddEventListener("click", func(dom.Object) { v.graph.createNode(n) })

		for _, p := range t.Panels {
			m := p.Name
			doc.ElementByID("node-"+n+"-"+m+"-link").
				AddEventListener("click", func(dom.Object) { v.selectedItem.(*Node).nc.ShowPartSubpanel(m) })
		}

	}
	return nil
}

func (v *View) setError(err string) {
	// TODO
	log.Print(err)
}

func (v *View) clearError() {
	// TODO
}

func (v *View) diagramCursorPos(e dom.Object) (x, y float64) {
	bcr := v.diagram.Call("getBoundingClientRect")
	x = e.Get("clientX").Float() - bcr.Get("left").Float()
	y = e.Get("clientY").Float() - bcr.Get("top").Float()
	return
}

func (v *View) diagramMouseDown(e dom.Object) {
	if v.selectedItem == nil {
		return
	}
	v.selectedItem.loseFocus(e)
	v.graph.gc.GainFocus()
	e.Call("stopPropagation")
}

func (v *View) diagramMouseMove(e dom.Object) {
	if v.dragItem == nil {
		return
	}
	v.dragItem.drag(e)
	e.Call("stopPropagation")
}

func (v *View) diagramMouseUp(e dom.Object) {
	if v.dragItem == nil {
		return
	}
	v.dragItem.drag(e)
	v.dragItem.drop(e)
	v.dragItem = nil
	e.Call("stopPropagation")
}

// selecter makes an onclick handler for a selectable.
func (v *View) selecter(s selectable) func(dom.Object) {
	return func(e dom.Object) {
		if v.selectedItem != nil {
			v.selectedItem.loseFocus(e)
		}
		v.selectedItem = s
		s.gainFocus(e)
		e.Call("stopPropagation")
	}
}

func (v *View) saveSelected(e dom.Object) {
	if v.selectedItem == nil {
		return
	}
	v.selectedItem.save(e)
}

func (v *View) deleteSelected(e dom.Object) {
	if v.selectedItem == nil {
		return
	}
	v.selectedItem.delete(e)
}

// draggable is anything that can be dragged on the canvas/SVG.
type draggable interface {
	drag(dom.Object)
	drop(dom.Object)
}

// selectable is anything that can be selected on the canvas/SVG.
type selectable interface {
	gainFocus(dom.Object)
	loseFocus(dom.Object)
	delete(dom.Object)
	save(dom.Object)
}
