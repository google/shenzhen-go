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
)

func TestGraphRefreshFromEmpty(t *testing.T) {
	v := makeFakeView() // Calls v.graph.refresh

	if v.graph.Channels == nil {
		t.Fatal("v.graph.Channels = nil, want non-nil map")
	}
	if v.graph.Nodes == nil {
		t.Fatal("v.graph.Nodes = nil, want non-nil map")
	}
	node1 := v.graph.Nodes["Node 1"]
	if node1 == nil {
		t.Fatal("v.graph.Nodes[Node 1] = nil, want non-nil node")
	}
	if got, want := len(node1.Inputs), 1; got != want {
		t.Errorf("len(Nodes[Node 1].Inputs) = %d, want %d", got, want)
	}
	if got, want := len(node1.Outputs), 2; got != want {
		t.Errorf("len(Nodes[Node 1].Outputs) = %d, want %d", got, want)
	}
	if got, want := len(node1.AllPins), 3; got != want {
		t.Errorf("len(Nodes[Node 1].AllPins) = %d, want %d", got, want)
	}

	if got, want := node1.TextBox.TextNode.Get("nodeValue").String(), "Node 1"; got != want {
		t.Errorf("Node 1 text = %q, want %q", got, want)
	}
}
