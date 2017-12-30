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
	"github.com/google/shenzhen-go/model"
	pb "github.com/google/shenzhen-go/proto/js"
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

type partEditor struct {
	Links  jsutil.Element
	Panels map[string]jsutil.Element
}

// View caches the top-level objects for managing the UI
type View struct {
	Document jsutil.Document     // Global document object
	Client   pb.ShenzhenGoClient // gRPC-Web client
	Diagram  *Diagram            // The LHS panel
	Graph    *Graph              // SVG elements in the LHS panel

	// RHS panels
	CurrentRHSPanel      jsutil.Element
	GraphPropertiesPanel jsutil.Element
	NodePropertiesPanel  jsutil.Element

	// Graph properties panel inputs
	graphNameElement        jsutil.Element
	graphPackagePathElement jsutil.Element
	graphIsCommandElement   jsutil.Element

	// Node properties subpanels and inputs
	nodeMetadataSubpanel  jsutil.Element
	nodeCurrentSubpanel   jsutil.Element
	nodeNameInput         jsutil.Element
	nodeEnabledInput      jsutil.Element
	nodeMultiplicityInput jsutil.Element
	nodeWaitInput         jsutil.Element
	nodePartEditors       map[string]*partEditor
}

// Setup connects to elements in the DOM.
func Setup(doc jsutil.Document, client pb.ShenzhenGoClient, initialJSON string) error {
	v := &View{
		Client:   client,
		Document: doc,

		Diagram: &Diagram{
			Element: doc.ElementByID("diagram"),
		},

		GraphPropertiesPanel: doc.ElementByID("graph-properties"),
		NodePropertiesPanel:  doc.ElementByID("node-properties"),
		CurrentRHSPanel:      doc.ElementByID("graph-properties"),

		graphNameElement:        doc.ElementByID("graph-prop-name"),
		graphPackagePathElement: doc.ElementByID("graph-prop-package-path"),
		graphIsCommandElement:   doc.ElementByID("graph-prop-is-command"),

		nodeMetadataSubpanel:  doc.ElementByID("node-metadata-panel"),
		nodeCurrentSubpanel:   doc.ElementByID("node-metadata-panel"),
		nodeNameInput:         doc.ElementByID("node-name"),
		nodeEnabledInput:      doc.ElementByID("node-enabled"),
		nodeMultiplicityInput: doc.ElementByID("node-multiplicity"),
		nodeWaitInput:         doc.ElementByID("node-wait"),
		nodePartEditors:       make(map[string]*partEditor, len(model.PartTypes)),
	}

	v.Diagram.errLabel = newTextBox(v, "", errTextStyle, errRectStyle, 0, 0, 0, 32)
	v.Diagram.errLabel.hide()

	g, err := loadGraph(v, initialJSON)
	if err != nil {
		return err
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
		AddEventListener("click", v.Diagram.saveSelected)
	doc.ElementByID("node-delete-link").
		AddEventListener("click", v.Diagram.deleteSelected)

	doc.ElementByID("node-metadata-link").
		AddEventListener("click", func(*js.Object) {
			v.Diagram.selectedItem.(*Node).showSubPanel(v.nodeMetadataSubpanel)
		})

	for n, t := range model.PartTypes {
		doc.ElementByID("node-new-link:"+n).
			AddEventListener("click", func(*js.Object) { v.Graph.createNode(n) })
		p := make(map[string]jsutil.Element, len(t.Panels))
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
					func(*js.Object) {
						v.Diagram.selectedItem.(*Node).showSubPanel(p)
					})
		}
	}

	return nil
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
