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

//+build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
)

const (
	activeColour = "#09f"
	normalColour = "#000"
	nodeStyle    = "fill: #eff; stroke: #355; stroke-width:1"
)

var (
	diagramSVG = js.Global.Get("document").Call("getElementById", "diagram")

	graph = &Graph{
		Nodes: []Node{
			{
				Name: "Hello, yes",
				Inputs: []Pin{
					{Name: "foo", Type: "int"},
				},
				Outputs: []Pin{
					{Name: "bar", Type: "string"},
				},
			},
			{
				Name: "this is dog",
				Inputs: []Pin{
					{Name: "baz", Type: "string"},
				},
				Outputs: []Pin{
					{Name: "qux", Type: "float64"},
				},
			},
		},
	}
)

type Pin struct {
	Name, Type string
}

type Node struct {
	Name            string
	Inputs, Outputs []Pin
}

type Graph struct {
	Nodes []Node
}

func main() {

}
