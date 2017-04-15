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
	document  = js.Global.Get("document")
	graphPath = js.Global.Get("graphPath").String()

	client api.Interface
)

func main() {
	apiURL := js.Global.Get("apiURL").String()
	if apiURL == "" {
		log.Fatalf("Couldn't find API URL")
	}
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
