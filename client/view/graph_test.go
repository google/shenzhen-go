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

	"github.com/google/shenzhen-go/jsutil"
	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/model/parts"
	"github.com/google/shenzhen-go/model/pin"
)

func TestGraphRefresh(t *testing.T) {
	doc := jsutil.MakeFakeDocument()
	v := &View{
		Document: doc,
		Diagram: &Diagram{
			Element: doc.MakeSVGElement("svg"),
		},
		Graph: &Graph{
			Graph: &model.Graph{
				Nodes: map[string]*model.Node{
					"Node 1": {
						Name:         "Node 1",
						Multiplicity: 1,
						Part: parts.NewCode(nil, "", "", "", pin.Map{
							"output": {
								Name:      "output",
								Direction: pin.Output,
								Type:      "int",
							},
						}),
					},
				},
			},
		},
	}
	v.Diagram.View = v
	v.Graph.View = v
	v.Graph.refresh()

	if v.Graph.Channels == nil {
		t.Error("g.Channels = nil, want non-nil map")
	}
	if v.Graph.Nodes == nil {
		t.Error("g.Nodes = nil, want non-nil map")
	}
	// TODO: inspect more state
}
