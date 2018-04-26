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
}

// Setup connects to elements in the DOM.
func Setup(doc dom.Document, gc GraphController) {
	v := &View{
		doc:     doc,
		diagram: doc.ElementByID("diagram"),
	}

	v.graph = &Graph{
		gc:     gc,
		doc:    doc,
		view:   v,
		errors: v,
	}
	v.graph.MakeElements(doc, v.diagram)

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
}

func (v *View) setError(err string) {
	// TODO
	log.Print(err)
}

func (v *View) clearError() {
	// TODO
	log.Print("*View.clearError")
}

func (v *View) createChannel(p *Pin) error {
	log.Print("*View.createChannel")

	cc, err := v.graph.gc.CreateChannel(p.pc)
	if err != nil {
		return err
	}
	ch := &Channel{
		cc:      cc,
		view:    v,
		errors:  v,
		graph:   v.graph,
		Pins:    make(map[*Pin]*Route),
		created: false,
	}
	ch.MakeElements(v.doc, v.diagram)
	ch.addPin(p)
	v.graph.Channels[cc.Name()] = ch
	ch.layout(nil)
	return nil
}

func (v *View) diagramCursorPos(e dom.Object) (x, y float64) {
	bcr := v.diagram.Call("getBoundingClientRect")
	x = e.Get("clientX").Float() - bcr.Get("left").Float()
	y = e.Get("clientY").Float() - bcr.Get("top").Float()
	return x, y
}

func (v *View) dragStarter(d dragStarter) func(dom.Object) {
	return func(e dom.Object) {
		e.Call("preventDefault")
		if dr, ok := d.(draggable); ok {
			v.dragItem = dr
		}
		d.dragStart(v.diagramCursorPos(e))
		e.Call("stopPropagation")
	}
}

func (v *View) diagramMouseDown(e dom.Object) {
	defer e.Call("stopPropagation")
	if v.selectedItem == nil {
		return
	}
	v.selectedItem.loseFocus()
	v.graph.gc.GainFocus()
}

func (v *View) diagramMouseMove(e dom.Object) {
	defer e.Call("stopPropagation")
	if v.dragItem == nil {
		return
	}
	v.dragItem.drag(v.diagramCursorPos(e))
}

func (v *View) diagramMouseUp(e dom.Object) {
	defer e.Call("stopPropagation")
	if v.dragItem == nil {
		return
	}
	v.dragItem.drag(v.diagramCursorPos(e))
	v.dragItem.drop()
	v.dragItem = nil
}

// selecter makes an onclick handler for a selectable.
func (v *View) selecter(s selectable) func(dom.Object) {
	return func(e dom.Object) {
		defer e.Call("stopPropagation")
		if v.selectedItem != nil {
			v.selectedItem.loseFocus()
		}
		v.selectedItem = s
		s.gainFocus()
	}
}

func (v *View) saveSelected(e dom.Object) {
	if v.selectedItem == nil {
		return
	}
	v.selectedItem.save()
}

func (v *View) deleteSelected(e dom.Object) {
	if v.selectedItem == nil {
		return
	}
	v.selectedItem.delete()
}

type dragStarter interface {
	dragStart(diagramX, diagramY float64)
}

// draggable is anything that can be dragged on the canvas/SVG.
type draggable interface {
	drag(diagramX, diagramY float64)
	drop()
}

// selectable is anything that can be selected on the canvas/SVG.
type selectable interface {
	gainFocus()
	loseFocus()
	delete()
	save()
}

type errorViewer interface {
	setError(string)
	clearError()
}
