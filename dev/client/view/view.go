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
	"github.com/google/shenzhen-go/dev/dom"
	"github.com/google/shenzhen-go/dev/model"
	pb "github.com/google/shenzhen-go/dev/proto/js"
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

type partEditor struct {
	Links  dom.Element
	Panels map[string]dom.Element
}

// View caches the top-level objects for managing the UI.
type View struct {
	doc     dom.Document // Global document object
	diagram *Diagram     // The LHS panel
	graph   *Graph       // SVG elements in the LHS panel

	client pb.ShenzhenGoClient

	// RHS panels
	CurrentRHSPanel      dom.Element
	GraphPropertiesPanel dom.Element
	NodePropertiesPanel  dom.Element

	// Node properties subpanels and inputs
	nodeMetadataSubpanel  dom.Element
	nodeCurrentSubpanel   dom.Element
	nodeNameInput         dom.Element
	nodeEnabledInput      dom.Element
	nodeMultiplicityInput dom.Element
	nodeWaitInput         dom.Element
	nodePartEditors       map[string]*partEditor
}

// Setup connects to elements in the DOM.
func Setup(doc dom.Document, client pb.ShenzhenGoClient, gc GraphController) error {
	v := &View{
		doc:    doc,
		client: client,

		GraphPropertiesPanel: doc.ElementByID("graph-properties"),
		NodePropertiesPanel:  doc.ElementByID("node-properties"),
		CurrentRHSPanel:      doc.ElementByID("graph-properties"),

		nodeMetadataSubpanel:  doc.ElementByID("node-metadata-panel"),
		nodeCurrentSubpanel:   doc.ElementByID("node-metadata-panel"),
		nodeNameInput:         doc.ElementByID("node-name"),
		nodeEnabledInput:      doc.ElementByID("node-enabled"),
		nodeMultiplicityInput: doc.ElementByID("node-multiplicity"),
		nodeWaitInput:         doc.ElementByID("node-wait"),
		nodePartEditors:       make(map[string]*partEditor, len(model.PartTypes)),
	}

	v.diagram = &Diagram{
		Element:  doc.ElementByID("diagram"),
		View:     v,
		errLabel: &TextBox{Margin: 20, TextOffsetY: 5},
	}
	v.diagram.errLabel.
		MakeElements(doc, v.diagram).
		SetTextStyle(errTextStyle).
		SetRectStyle(errRectStyle).
		SetHeight(32)

	v.graph = &Graph{
		gc:   gc,
		doc:  doc,
		view: v,
	}
	v.graph.makeElements(doc, v.diagram)
	v.graph.refresh()

	v.diagram.
		AddEventListener("mousedown", v.diagram.mouseDown).
		AddEventListener("mousemove", v.diagram.mouseMove).
		AddEventListener("mouseup", v.diagram.mouseUp)

	doc.ElementByID("graph-save").
		AddEventListener("click", v.graph.save)
	doc.ElementByID("graph-properties-save").
		AddEventListener("click", v.graph.saveProperties)

	doc.ElementByID("node-save-link").
		AddEventListener("click", v.diagram.saveSelected)
	doc.ElementByID("node-delete-link").
		AddEventListener("click", v.diagram.deleteSelected)

	doc.ElementByID("node-metadata-link").
		AddEventListener("click", func(dom.Object) {
			v.diagram.selectedItem.(*Node).showSubPanel(v.nodeMetadataSubpanel)
		})

	for n, t := range gc.PartTypes() {
		doc.ElementByID("node-new-link:"+n).
			AddEventListener("click", func(dom.Object) { v.graph.createNode(n) })
		p := make(map[string]dom.Element, len(t.Panels))
		for _, d := range t.Panels {
			p[d.Name] = doc.ElementByID("node-" + n + "-" + d.Name + "-panel")
		}
		v.nodePartEditors[n] = &partEditor{
			Links:  doc.ElementByID("node-" + n + "-links"),
			Panels: p,
		}
	}
	for n, e := range v.nodePartEditors {
		for m, p := range e.Panels {
			p := p
			doc.ElementByID("node-"+n+"-"+m+"-link").
				AddEventListener("click",
					func(dom.Object) {
						v.diagram.selectedItem.(*Node).showSubPanel(p)
					})
		}
	}

	return nil
}

// ShowRHSPanel hides any existing panel and shows the given element as the panel.
func (v *View) ShowRHSPanel(p dom.Element) {
	if p == v.CurrentRHSPanel {
		return
	}
	v.CurrentRHSPanel.Hide()
	v.CurrentRHSPanel = p.Show()
}

func (v *View) setError(err string) {
	v.diagram.setError(err, 0, 0)
}

func (v *View) clearError() {
	v.diagram.clearError()
}
