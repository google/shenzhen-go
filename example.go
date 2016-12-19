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

package main

// TODO: Have this as a JSON file loaded at runtime.

import (
	"shenzhen-go/graph"
	"shenzhen-go/parts"
)

var exampleGraph = &graph.Graph{
	Name:        "Example",
	PackageName: "example",
	PackagePath: "example", // == $GOPATH/src/example
	Imports: []string{
		"fmt",
	},
	Nodes: map[string]*graph.Node{
		"Generate integers ≥ 2": {
			Name: "Generate integers ≥ 2",
			Part: &parts.Code{
				Code: `for i:= 2; i<100; i++ {
	raw <- i
}
close(raw)`,
			},
			Wait: true,
		},
		"Filter divisible by 2": {
			Name: "Filter divisible by 2",
			Part: &parts.Code{Code: `for n := range raw {
	if n > 2 && n % 2 == 0 {
		continue
	}
	div2 <- n
}
close(div2)`,
			},
			Wait: true,
		},
		"Filter divisible by 3": {
			Name: "Filter divisible by 3",
			Part: &parts.Code{Code: `for n := range div2 {
	if n > 3 && n % 3 == 0 {
		continue
	}
	out <- n
}
close(out)`,
			},
			Wait: true,
		},
		"Print output": {
			Name: "Print output",
			Part: &parts.Code{Code: `for n := range out {
	fmt.Println(n)
}`,
			},
			Wait: true,
		},
	},
	Channels: map[string]*graph.Channel{
		"raw": {
			Name: "raw",
			Type: "int",
			Cap:  0,
		},
		"div2": {
			Name: "div2",
			Type: "int",
			Cap:  0,
		},
		"out": {
			Name: "out",
			Type: "int",
			Cap:  0,
		},
	},
}

func init() {
	for _, n := range exampleGraph.Nodes {
		n.Refresh(exampleGraph)
	}
}
