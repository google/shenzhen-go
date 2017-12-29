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
	"testing"

	"github.com/google/shenzhen-go/jsutil"
	"github.com/google/shenzhen-go/model"
)

func TestGraphRefresh(t *testing.T) {
	theDocument = jsutil.MakeFakeDocument()
	theDiagram = &Diagram{Element: jsutil.MakeFakeElement("svg", jsutil.SVGNamespaceURI)}
	g := &Graph{
		Graph: &model.Graph{},
	}
	g.refresh()

	if g.Channels == nil {
		t.Error("g.Channels = nil, want non-nil map")
	}
	if g.Nodes == nil {
		t.Error("g.Nodes = nil, want non-nil map")
	}
	// TODO: inspect more state
}
