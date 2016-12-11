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

var exampleGraph = Graph{
	Name:        "Example",
	PackageName: "example",
	PackagePath: "example", // == $GOPATH/src/example
	Imports: []string{
		"fmt",
	},
	Nodes: map[string]*Node{
		"Generate integers ≥ 2": {
			Name: "Generate integers ≥ 2",
			Code: `for i:= 2; i<100; i++ {
	raw <- i
}
close(raw)`,
			Wait: true,
		},
		"Filter divisible by 2": {
			Name: "Filter divisible by 2",
			Code: `for n := range raw {
	if n > 2 && n % 2 == 0 {
		continue
	}
	div2 <- n
}
close(div2)`,
			Wait: true,
		},
		"Filter divisible by 3": {
			Name: "Filter divisible by 3",
			Code: `for n := range div2 {
	if n > 3 && n % 3 == 0 {
		continue
	}
	out <- n
}
close(out)`,
			Wait: true,
		},
		"Print output": {
			Name: "Print output",
			Code: `for n := range out {
	fmt.Println(n)
}`,
			Wait: true,
		},
	},
	Channels: map[string]*Channel{
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
		n.updateChans(exampleGraph.Channels)
	}
}
