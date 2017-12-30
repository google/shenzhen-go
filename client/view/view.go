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
	"github.com/google/shenzhen-go/jsutil"
	pb "github.com/google/shenzhen-go/proto/js"
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

// View caches the top-level objects for managing the UI
type View struct {
	Document jsutil.Document
	Client   pb.ShenzhenGoClient
	Diagram  *Diagram // The left panel
	Graph    *Graph

	CurrentRHSPanel      jsutil.Element
	GraphPropertiesPanel jsutil.Element
	NodePropertiesPanel  jsutil.Element
}

// Setup connects to elements in the DOM.
func Setup(doc jsutil.Document, client pb.ShenzhenGoClient, initialJSON string) {
	v := &view.View{
		Client:   client,
		Document: doc,

		Diagram: &view.Diagram{Element: doc.ElementByID("diagram")},

		GraphPropertiesPanel: doc.ElementByID("graph-properties"),
		NodePropertiesPanel:  doc.ElementByID("node-properties"),
		RHSPanel:             doc.ElementByID("graph-properties"),
	}

	v.Diagram.errLabel = newTextBox("", errTextStyle, errRectStyle, 0, 0, 0, 32)
	v.Diagram.errLabel.hide()

	g, err := loadGraph(initialJSON)
	if err != nil {
		log.Fatalf("Couldn't load graph: %v", err)
	}

	v.Diagram.
		AddEventListener("mousedown", v.Diagram.mouseDown).
		AddEventListener("mousemove", v.Diagram.mouseMove).
		AddEventListener("mouseup", v.Diagram.mouseUp)

	doc.ElementByID("graph-save").
		AddEventListener("click", g.save)
	doc.ElementByID("graph-properties-save").
		AddEventListener("click", g.saveProperties)

	doc.ElementByID("node-save-link").
		AddEventListener("click", theDiagram.saveSelected)
	doc.ElementByID("node-delete-link").
		AddEventListener("click", theDiagram.deleteSelected)

	doc.ElementByID("node-metadata-link").
		AddEventListener("click", func(*js.Object) {
			theDiagram.selectedItem.(*Node).showSubPanel(nodeMetadataSubpanel)
		})

	for n, e := range nodePartEditors {
		for m, p := range e.Panels {
			p := p
			doc.ElementByID(fmt.Sprintf("node-%s-%s-link", n, m)).
				AddEventListener("click",
					func(*js.Object) {
						theDiagram.selectedItem.(*Node).showSubPanel(p)
					})
		}
	}
}

// ShowRHSPanel hides any existing panel and shows the given element as the panel.
func (v *View) ShowRHSPanel(p jsutil.Element) {
	if p == v.CurrentRHSPanel {
		return
	}
	v.CurrentRHSPanel.Get("style").Set("display", "none")
	v.CurrentRHSPanel = p
	v.CurrentRHSPanel.Get("style").Set("display", nil)
}
