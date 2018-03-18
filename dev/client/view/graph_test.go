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

package view

import (
	"testing"

	"github.com/google/shenzhen-go/dev/dom"
	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/parts"
	"github.com/google/shenzhen-go/dev/model/pin"
)

func TestGraphRefreshFromEmpty(t *testing.T) {
	doc := dom.MakeFakeDocument()
	v := &View{Document: doc}
	v.Diagram = &Diagram{
		View:    v,
		Element: doc.MakeSVGElement("svg"),
	}
	v.Graph = &Graph{
		View: v,
		Graph: &model.Graph{
			Nodes: map[string]*model.Node{
				"Node 1": {
					Name:         "Node 1",
					Multiplicity: 1,
					Part: parts.NewCode(nil, "", "", "", pin.Map{
						"input": {
							Name:      "input",
							Direction: pin.Input,
							Type:      "int",
						},
						"output": {
							Name:      "output",
							Direction: pin.Output,
							Type:      "int",
						},
						"output 2": {
							Name:      "output 2",
							Direction: pin.Output,
							Type:      "int",
						},
					}),
				},
			},
		},
	}
	v.Graph.refresh()

	if v.Graph.Channels == nil {
		t.Error("g.Channels = nil, want non-nil map")
	}
	if v.Graph.Nodes == nil {
		t.Error("g.Nodes = nil, want non-nil map")
	}
	if got, want := len(v.Graph.Nodes["Node 1"].Inputs), 1; got != want {
		t.Errorf("len(Nodes[Node 1].Inputs) = %d, want %d", got, want)
	}
	if got, want := len(v.Graph.Nodes["Node 1"].Outputs), 2; got != want {
		t.Errorf("len(Nodes[Node 1].Outputs) = %d, want %d", got, want)
	}
	if got, want := len(v.Graph.Nodes["Node 1"].AllPins), 3; got != want {
		t.Errorf("len(Nodes[Node 1].AllPins) = %d, want %d", got, want)
	}

	if got, want := v.Graph.Nodes["Node 1"].box.textNode.Get("wholeText").String(), "Node 1"; got != want {
		t.Errorf("Node 1 text = %q, want %q", got, want)
	}
}
