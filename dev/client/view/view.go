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

	"github.com/google/shenzhen-go/dev/dom"
	"github.com/google/shenzhen-go/dev/model"
)

const (
	hoverTipOffset = Point(complex(8, 8))
	hoverTipHeight = 30
	pinRadius      = 5
	snapDist       = 12
)

// View caches the top-level objects for managing the UI.
type View struct {
	doc      dom.Document // Global document object
	diagram  dom.Element  // The LHS panel
	graph    *Graph       // manages most SVG elements inside diagram
	hoverTip *TextBox

	dragItem     dragger  // nil if nothing is being dragged
	selectedItem selecter // nil if nothing is selected
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
	v.hoverTip = (&TextBox{Margin: nodeBoxMargin}).
		MakeElements(doc, v.diagram).
		SetHeight(hoverTipHeight)
	v.hoverTip.Hide()

	v.selectedItem = v.graph

	v.diagram.
		AddEventListener("mousedown", v.selecter(v.graph)).
		AddEventListener("mousemove", v.diagramMouseMove).
		AddEventListener("mouseup", v.diagramMouseUp)

	doc.ElementByID("graph-save").
		AddEventListener("click", v.graph.save)
	doc.ElementByID("graph-revert").
		AddEventListener("click", v.graph.revert)
	doc.ElementByID("graph-generate").
		AddEventListener("click", v.graph.generate)
	doc.ElementByID("graph-build").
		AddEventListener("click", v.graph.build)
	doc.ElementByID("graph-install").
		AddEventListener("click", v.graph.install)
	doc.ElementByID("graph-run").
		AddEventListener("click", v.graph.run)

	doc.ElementByID("preview-go-link").
		AddEventListener("click", func(dom.Object) { gc.PreviewGo() })
	doc.ElementByID("preview-raw-go-link").
		AddEventListener("click", func(dom.Object) { gc.PreviewRawGo() })
	doc.ElementByID("preview-json-link").
		AddEventListener("click", func(dom.Object) { gc.PreviewJSON() })

	doc.ElementByID("help-licenses-link").
		AddEventListener("click", func(dom.Object) { gc.HelpLicenses() })
	doc.ElementByID("help-about-link").
		AddEventListener("click", func(dom.Object) { gc.HelpAbout() })

	doc.ElementByID("graph-prop-name").
		AddEventListener("change", v.graph.commit)
	doc.ElementByID("graph-prop-package-path").
		AddEventListener("change", v.graph.commit)
	doc.ElementByID("graph-prop-is-command").
		AddEventListener("change", v.graph.commit)

	doc.ElementByID("channel-name").
		AddEventListener("change", v.commitSelected)
	doc.ElementByID("channel-capacity").
		AddEventListener("change", v.commitSelected)

	doc.ElementByID("channel-delete-link").
		AddEventListener("click", v.deleteSelected)

	doc.ElementByID("node-name").
		AddEventListener("change", v.commitSelected)
	doc.ElementByID("node-enabled").
		AddEventListener("change", v.commitSelected)
	doc.ElementByID("node-multiplicity").
		AddEventListener("change", v.commitSelected)
	doc.ElementByID("node-wait").
		AddEventListener("change", v.commitSelected)

	// TODO(josh): reinstate Clone and Convert-To-Code links
	doc.ElementByID("node-delete-link").
		AddEventListener("click", v.deleteSelected)

	doc.ElementByID("node-metadata-link").
		AddEventListener("click", func(dom.Object) {
			v.selectedItem.(*Node).nc.ShowMetadataSubpanel()
		})

	for n, t := range model.PartTypes {
		doc.ElementByID("node-new-link:"+n).
			AddEventListener("click", func(dom.Object) { go v.graph.reallyCreateNode(n) }) // Don't block in callback

		for _, p := range t.Panels {
			m := p.Name
			doc.ElementByID("node-"+n+"-"+m+"-link").
				AddEventListener("click", func(dom.Object) { v.selectedItem.(*Node).nc.ShowPartSubpanel(m) })
		}

	}
}

func (v *View) showHoverTip(event dom.Object, tip string) {
	v.hoverTip.
		SetText(tip).
		RecomputeWidth().
		MoveTo(v.diagramCursorPos(event) + hoverTipOffset).
		Show()
	event.Call("stopPropagation")
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
		cc:     cc,
		view:   v,
		errors: v,
		graph:  v.graph,
		Pins:   make(map[*Pin]*Route),
	}
	ch.MakeElements(v.doc, v.diagram)
	ch.addPin(p)
	v.graph.Channels[cc.Name()] = ch
	ch.layout(nil)
	return nil
}

func (v *View) diagramCursorPos(e dom.Object) Point {
	bcr := v.diagram.Call("getBoundingClientRect")
	x := e.Get("clientX").Float() - bcr.Get("left").Float()
	y := e.Get("clientY").Float() - bcr.Get("top").Float()
	return Pt(x, y)
}

func (v *View) dragStarter(d dragStarter) func(dom.Object) {
	return func(e dom.Object) {
		e.Call("preventDefault")
		if dr, ok := d.(dragger); ok {
			v.dragItem = dr
		}
		d.dragStart(v.diagramCursorPos(e))
		e.Call("stopPropagation")
	}
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

// changeSelection switches the selected item, calling loseFocus and gainFocus
// as necessary.
func (v *View) changeSelection(s selecter) {
	v.selectedItem.loseFocus()
	v.selectedItem = s
	s.gainFocus()
}

// selecter makes an onclick handler for a selectable.
func (v *View) selecter(s selecter) func(dom.Object) {
	return func(e dom.Object) {
		v.changeSelection(s)
		e.Call("stopPropagation")
	}
}

func (v *View) commitSelected(e dom.Object) {
	s, _ := v.selectedItem.(commitDeleter)
	if s == nil {
		return
	}
	s.commit()
}

func (v *View) deleteSelected(e dom.Object) {
	s, _ := v.selectedItem.(commitDeleter)
	if s == nil {
		return
	}
	s.delete()
	v.graph.gainFocus()
}

// dragStarter is anything that can start a drag action
type dragStarter interface {
	dragStart(Point)
}

// dragger is anything that can be dragged on the canvas/SVG.
type dragger interface {
	drag(Point)
	drop()
}

// selecter is anything that can be selected on the canvas/SVG.
type selecter interface {
	gainFocus()
	loseFocus()
}

// commitDeleter is anything that can be "saved" or "deleted".
type commitDeleter interface {
	commit()
	delete()
}

type errorViewer interface {
	setError(string)
	clearError()
}
