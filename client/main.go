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

// +build js

package main

import (
	"fmt"
	"log"

	"github.com/google/shenzhen-go/jsutil"
	pb "github.com/google/shenzhen-go/proto"
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
	graphPath = jsutil.MustGetGlobal("graphPath").String()

	graphPropertiesPanel = jsutil.MustGetElement("graph-properties")
	nodePropertiesPanel  = jsutil.MustGetElement("node-properties")
	rhsPanel             = graphPropertiesPanel

	theClient  pb.ShenzhenGoClient
	theDiagram *Diagram
	theGraph   *Graph
)

func showRHSPanel(p *jsutil.Element) {
	if p == rhsPanel {
		return
	}
	rhsPanel.Get("style").Set("display", "none")
	rhsPanel = p
	rhsPanel.Get("style").Set("display", nil)
}

func main() {
	theClient = pb.NewShenzhenGoClient(jsutil.MustGetGlobal("apiURL").String())

	theDiagram = &Diagram{Element: jsutil.MustGetElement("diagram")}
	theDiagram.errLabel = newTextBox("", errTextStyle, errRectStyle, 0, 0, 0, 32)
	theDiagram.errLabel.hide()

	g, err := loadGraph(jsutil.MustGetGlobal("GraphJSON").String())
	if err != nil {
		log.Fatalf("Couldn't load graph: %v", err)
	}
	theGraph = g

	theDiagram.
		AddEventListener("mousedown", theDiagram.mouseDown).
		AddEventListener("mousemove", theDiagram.mouseMove).
		AddEventListener("mouseup", theDiagram.mouseUp)

	jsutil.MustGetElement("graph-save").
		AddEventListener("click", g.save)
	jsutil.MustGetElement("graph-properties-save").
		AddEventListener("click", g.saveProperties)

	jsutil.MustGetElement("node-save-link").
		AddEventListener("click", theDiagram.saveSelected)
	jsutil.MustGetElement("node-delete-link").
		AddEventListener("click", theDiagram.deleteSelected)

	jsutil.MustGetElement("node-metadata-link").
		AddEventListener("click", func(*js.Object) {
			theDiagram.selectedItem.(*Node).showSubPanel(nodeMetadataSubpanel)
		})

	for n, e := range nodePartEditors {
		for m, p := range e.Panels {
			p := p
			jsutil.MustGetElement(fmt.Sprintf("node-%s-%s-link", n, m)).
				AddEventListener("click",
					func(*js.Object) {
						theDiagram.selectedItem.(*Node).showSubPanel(p)
					})
		}
	}
}
