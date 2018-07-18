// Copyright 2016 Google Inc.
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

package view

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/google/shenzhen-go/dev/model"
)

var graphEditorTemplate = template.Must(template.New("graphEditor").Parse(string(templateResources["templates/graph.html"])))

type editorInput struct {
	Params              *Params
	Graph               *model.Graph
	GraphJSON           string
	PartTypes           map[string]*model.PartType
	PartTypesByCategory map[string]map[string]*model.PartType
	Licenses            struct {
		ShenzhenGo string
		Ace        string
		Hterm      string
	}
}

// Graph displays a graph.
func Graph(w http.ResponseWriter, g *model.Graph, params *Params) {
	gj, err := json.Marshal(g)
	if err != nil {
		log.Printf("Could not execute graph editor template: %v", err)
		http.Error(w, "Could not execute graph editor template", http.StatusInternalServerError)
		return
	}

	d := &editorInput{
		Params:              params,
		Graph:               g,
		GraphJSON:           string(gj),
		PartTypes:           model.PartTypes,
		PartTypesByCategory: model.PartTypesByCategory,
	}
	d.Licenses.ShenzhenGo = string(miscResources["misc/LICENSE"])
	d.Licenses.Ace = string(jsResources["js/ace/LICENSE"])
	d.Licenses.Hterm = string(jsResources["js/hterm/LICENSE"])
	if err := graphEditorTemplate.Execute(w, d); err != nil {
		log.Printf("Could not execute graph editor template: %v", err)
		http.Error(w, "Could not execute graph editor template", http.StatusInternalServerError)
	}
}
