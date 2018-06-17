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
	"context"

	"github.com/google/shenzhen-go/dev/dom"
)

type fakeGraphController struct{}

func (c fakeGraphController) GainFocus()                         {}
func (c fakeGraphController) Nodes(f func(NodeController))       { f(fakeNodeController{}) }
func (c fakeGraphController) NumNodes() int                      { return 1 }
func (c fakeGraphController) Channels(f func(ChannelController)) {}
func (c fakeGraphController) NumChannels() int                   { return 0 }

func (c fakeGraphController) CreateChannel(...PinController) (ChannelController, error) {
	return nil, nil
}

func (c fakeGraphController) CreateNode(ctx context.Context, partType string) (NodeController, error) {
	return nil, nil
}

func (c fakeGraphController) Commit(ctx context.Context) error   { return nil }
func (c fakeGraphController) Save(ctx context.Context) error     { return nil }
func (c fakeGraphController) Revert(ctx context.Context) error   { return nil }
func (c fakeGraphController) Generate(ctx context.Context) error { return nil }
func (c fakeGraphController) Build(ctx context.Context) error    { return nil }
func (c fakeGraphController) Install(ctx context.Context) error  { return nil }
func (c fakeGraphController) Run(ctx context.Context) error      { return nil }
func (c fakeGraphController) PreviewGo()                         {}
func (c fakeGraphController) PreviewJSON()                       {}

type fakeNodeController struct{}

func (f fakeNodeController) Name() string             { return "Node 1" }
func (f fakeNodeController) Position() (x, y float64) { return 150, 150 }

func (f fakeNodeController) Pins(x func(PinController, string)) {
	x(fakePinController("input"), "nil")
	x(fakePinController("output"), "nil")
	x(fakePinController("output 2"), "nil")
}

func (f fakeNodeController) GainFocus() {}

func (f fakeNodeController) Delete(context.Context) error                        { return nil }
func (f fakeNodeController) Commit(context.Context) error                        { return nil }
func (f fakeNodeController) SetPosition(context.Context, float64, float64) error { return nil }
func (f fakeNodeController) ShowMetadataSubpanel()                               {}
func (f fakeNodeController) ShowPartSubpanel(string)                             {}

type fakePinController string

func (f fakePinController) Name() string     { return string(f) }
func (f fakePinController) Type() string     { return "int" }
func (f fakePinController) IsInput() bool    { return f == "input" }
func (f fakePinController) NodeName() string { return "Node 1" }

func makeFakeView() *View {
	doc := dom.MakeFakeDocument()
	v := &View{
		doc:     doc,
		diagram: doc.MakeSVGElement("svg"),
	}
	v.graph = &Graph{
		view: v,
		gc:   fakeGraphController{},
		doc:  doc,
	}
	v.graph.MakeElements(doc, v.diagram)
	return v
}
