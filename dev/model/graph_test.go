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

package model

import (
	"fmt"
	"strings"
	"testing"

	"gopkg.in/d4l3k/messagediff.v1"

	"github.com/google/shenzhen-go/dev/model/pin"
	"github.com/google/shenzhen-go/dev/source"
)

func init() {
	RegisterPartType("Fake", "Misc", &PartType{
		New: func() Part { return new(FakePart) },
	})
}

func TestLoadJSON(t *testing.T) {
	json := strings.NewReader(`{
	"nodes": {
		"foo": {
			"part_type": "Fake", 
			"part": {
				"pins": {
					"output": {
						"type": "int",
						"dir": "out"
					},
					"nc": {
						"type": "int",
						"dir": "in"
					}
				}
			},
			"connections": {
				"output": "baz",
				"ignored": "ignored"
			}
		},
		"bar": {
			"part_type": "Fake", 
			"part": {
				"pins": {
					"input": {
						"type": "int",
						"dir": "in"
					},
					"nc": {
						"type": "int",
						"dir": "out"
					}
				}
			},
			"connections": {
				"input": "baz",
				"ignored": "ignored"
			}
		}
	},
	"channels": {
		"baz": {
			"type": "int"
		}
	}
}`)
	got, err := LoadJSON(json, "filePath", "urlPath")
	if err != nil {
		t.Fatalf("LoadJSON() error = %v", err)
	}

	if got, want := got.FilePath, "filePath"; got != want {
		t.Errorf("LoadJSON().FilePath = %q, want %q", got, want)
	}
	if got, want := got.URLPath, "urlPath"; got != want {
		t.Errorf("LoadJSON().URLPath = %q, want %q", got, want)
	}

	wantNodes := map[string]*Node{
		"foo": {
			Name:         "foo",
			Multiplicity: "1",
			Part: &FakePart{nil, "", "", "", pin.Map{
				"output": {
					Name:      "output",
					Direction: pin.Output,
					Type:      "int",
				},
				"nc": {
					Name:      "nc",
					Type:      "int",
					Direction: pin.Input,
				},
			}},
			Connections: map[string]string{
				"output": "baz",
				"nc":     "nil",
			},
		},
		"bar": {
			Name:         "bar",
			Multiplicity: "1",
			Part: &FakePart{nil, "", "", "", pin.Map{
				"input": {
					Name:      "input",
					Direction: pin.Input,
					Type:      "int",
				},
				"nc": {
					Name:      "nc",
					Type:      "int",
					Direction: pin.Output,
				},
			}},
			Connections: map[string]string{
				"input": "baz",
				"nc":    "nil",
			},
		},
	}
	if diff, equal := messagediff.PrettyDiff(got.Nodes, wantNodes); !equal {
		t.Errorf("LoadJSON().Nodes diff (got -> want)\n%v", diff)
	}
	wantChans := map[string]*Channel{
		"baz": {
			Name: "baz",
			Pins: map[NodePin]struct{}{
				{Node: "foo", Pin: "output"}: {},
				{Node: "bar", Pin: "input"}:  {},
			},
		},
	}
	if diff, equal := messagediff.PrettyDiff(got.Channels, wantChans); !equal {
		t.Errorf("LoadJSON().Channels diff (got -> want)\n%v", diff)
	}
}

func TestInferTypesSimple(t *testing.T) {
	g := &Graph{
		FilePath:    "filepath",
		URLPath:     "urlpath",
		Name:        "basic inference",
		PackagePath: "package/path",
		IsCommand:   false,
		Nodes: map[string]*Node{
			"node 1": {
				Part: &FakePart{nil, "", "", "", pin.NewMap(&pin.Definition{
					Name:      "output",
					Type:      "int",
					Direction: pin.Output,
				})},
				Name:         "node 1",
				Enabled:      true,
				Multiplicity: "1",
				Wait:         true,
				Connections: map[string]string{
					"output": "bar",
				},
			},
			"node 2": {
				Part: &FakePart{nil, "", "", "", pin.NewMap(&pin.Definition{
					Name:      "input",
					Type:      "$T",
					Direction: pin.Input,
				})},
				Name:         "node 2",
				Enabled:      true,
				Multiplicity: "1",
				Wait:         true,
				Connections: map[string]string{
					"input": "bar",
				},
			},
		},
		Channels: map[string]*Channel{
			"bar": {
				Name:     "bar",
				Capacity: 0,
			},
		},
	}
	g.RefreshChannelsPins()

	if err := g.InferTypes(); err != nil {
		t.Fatalf("InferTypes() = error %v", err)
	}
	// bar should have type "int"
	if got, want := g.Channels["bar"].Type.String(), "int"; got != want {
		t.Errorf("Channels[bar].Type = %s, want %s", got, want)
	}
	// node 1.output should still have type "int"
	if got, want := g.Nodes["node 1"].PinTypes["output"].String(), "int"; got != want {
		t.Errorf("Nodes[node 1].PinTypes[output] = %s, want %s", got, want)
	}
	// node 2.input should have type "int"
	if got, want := g.Nodes["node 2"].PinTypes["input"].String(), "int"; got != want {
		t.Errorf("Nodes[node 2].PinTypes[input] = %s, want %s", got, want)
	}
	// node 1.TypeParams should be an empty map.
	if diff, equal := messagediff.PrettyDiff(g.Nodes["node 1"].TypeParams, map[string]*source.Type{}); !equal {
		t.Errorf("Nodes[node 1].TypeParams diff:\n%s", diff)
	}
	// node 2.TypeParams should give a value for $T.
	got := make(map[string]string)
	for k, t := range g.Nodes["node 2"].TypeParams {
		got[k] = t.String()
	}
	if diff, equal := messagediff.PrettyDiff(got, map[string]string{"$T": "int"}); !equal {
		t.Errorf("Nodes[node 2].TypeParams diff:\n%s", diff)
	}
}

func TestInferTypesNoChannel(t *testing.T) {
	g := &Graph{
		FilePath:    "filepath",
		URLPath:     "urlpath",
		Name:        "basic inference",
		PackagePath: "package/path",
		IsCommand:   false,
		Nodes: map[string]*Node{
			"node 1": {
				Part: &FakePart{nil, "", "", "", pin.NewMap(
					&pin.Definition{
						Name:      "input",
						Type:      "$A",
						Direction: pin.Input,
					},
					&pin.Definition{
						Name:      "output",
						Type:      "$B",
						Direction: pin.Output,
					})},
				Name:         "node 1",
				Enabled:      true,
				Multiplicity: "1",
				Wait:         true,
				Connections: map[string]string{
					"output": "bar",
				},
			},
		},
	}
	g.RefreshChannelsPins()

	if err := g.InferTypes(); err != nil {
		t.Fatalf("InferTypes() = error %v", err)
	}
	// node 1.output should have type "interface{}"
	if got, want := g.Nodes["node 1"].PinTypes["output"].String(), "interface{}"; got != want {
		t.Errorf("Nodes[node 1].PinTypes[output] = %s, want %s", got, want)
	}
	// node 1.input should have type "interface{}"
	if got, want := g.Nodes["node 1"].PinTypes["input"].String(), "interface{}"; got != want {
		t.Errorf("Nodes[node 2].PinTypes[input] = %s, want %s", got, want)
	}
	// node 1.TypeParams should give a value for $A and $B.
	got := make(map[string]string)
	for k, t := range g.Nodes["node 1"].TypeParams {
		got[k] = t.String()
	}
	want := map[string]string{
		"$A": "interface{}",
		"$B": "interface{}",
	}
	if diff, equal := messagediff.PrettyDiff(got, want); !equal {
		t.Errorf("Nodes[node 1].TypeParams diff:\n%s", diff)
	}
}

func TestInferTypesMapToMap(t *testing.T) {
	g := &Graph{
		FilePath:    "filepath",
		URLPath:     "urlpath",
		Name:        "map to map",
		PackagePath: "package/path",
		IsCommand:   false,
		Nodes: map[string]*Node{
			"node 1": {
				Part: &FakePart{nil, "", "", "", pin.NewMap(&pin.Definition{
					Name:      "output",
					Type:      "map[$K]int",
					Direction: pin.Output,
				})},
				Name:         "node 1",
				Enabled:      true,
				Multiplicity: "1",
				Wait:         true,
				Connections: map[string]string{
					"output": "bar",
				},
			},
			"node 2": {
				Part: &FakePart{nil, "", "", "", pin.NewMap(&pin.Definition{
					Name:      "input",
					Type:      "map[string]$V",
					Direction: pin.Input,
				})},
				Name:         "node 2",
				Enabled:      true,
				Multiplicity: "1",
				Wait:         true,
				Connections: map[string]string{
					"input": "bar",
				},
			},
		},
		Channels: map[string]*Channel{
			"bar": {
				Name:     "bar",
				Capacity: 0,
			},
		},
	}
	g.RefreshChannelsPins()

	if err := g.InferTypes(); err != nil {
		t.Fatalf("InferTypes() = error %v", err)
	}
	want := "map[string]int"
	if got := g.Channels["bar"].Type.String(); got != want {
		t.Errorf("Channels[bar].Type = %s, want %s", got, want)
	}
	if got := g.Nodes["node 1"].PinTypes["output"].String(); got != want {
		t.Errorf("Nodes[node 1].PinTypes[output] = %s, want %s", got, want)
	}
	if got := g.Nodes["node 2"].PinTypes["input"].String(); got != want {
		t.Errorf("Nodes[node 2].PinTypes[input] = %s, want %s", got, want)
	}
}

func TestInferTypes10Chain(t *testing.T) {
	g := &Graph{
		FilePath:    "filepath",
		URLPath:     "urlpath",
		Name:        "chain with 10 channels",
		PackagePath: "package/path",
		IsCommand:   false,
		Nodes: map[string]*Node{
			"node 0": {
				Part: &FakePart{nil, "", "", "", pin.NewMap(&pin.Definition{
					Name:      "output",
					Type:      "map[$K]int",
					Direction: pin.Output,
				})},
				Name:         "node 0",
				Enabled:      true,
				Multiplicity: "1",
				Wait:         true,
				Connections: map[string]string{
					"output": "chan0_1",
				},
			},
			"node 10": {
				Part: &FakePart{nil, "", "", "", pin.NewMap(&pin.Definition{
					Name:      "input",
					Type:      "map[string]$V",
					Direction: pin.Input,
				})},
				Name:         "node 10",
				Enabled:      true,
				Multiplicity: "1",
				Wait:         true,
				Connections: map[string]string{
					"input": "chan9_10",
				},
			},
		},
		Channels: map[string]*Channel{
			"chan0_1": {
				Name: "chan0_1",
			},
		},
	}
	for i := 1; i < 10; i++ {
		name := fmt.Sprintf("node %d", i)
		cname := fmt.Sprintf("chan%d_%d", i, i+1)
		g.Nodes[name] = &Node{
			Part: &FakePart{nil, "", "", "", pin.NewMap(
				&pin.Definition{
					Name:      "input",
					Type:      "$T",
					Direction: pin.Input,
				},
				&pin.Definition{
					Name:      "output",
					Type:      "$T",
					Direction: pin.Output,
				},
			)},
			Name:         name,
			Enabled:      true,
			Multiplicity: "1",
			Wait:         true,
			Connections: map[string]string{
				"input":  fmt.Sprintf("chan%d_%d", i-1, i),
				"output": cname,
			},
		}
		g.Channels[cname] = &Channel{
			Name: cname,
		}
	}
	g.RefreshChannelsPins()

	if err := g.InferTypes(); err != nil {
		t.Fatalf("InferTypes() = error %v", err)
	}
	want := "map[string]int"
	for _, c := range g.Channels {
		if got := c.Type.String(); got != want {
			t.Errorf("Channels[%s].Type = %s, want %s", c.Name, got, want)
		}
	}
	for _, n := range g.Nodes {
		if got := n.PinTypes["input"].String(); n.Name != "node 0" && got != want {
			t.Errorf("Nodes[%s].Type = %s, want %s", n.Name, got, want)
		}
		if got := n.PinTypes["output"].String(); n.Name != "node 10" && got != want {
			t.Errorf("Nodes[%s].Type = %s, want %s", n.Name, got, want)
		}
	}
}
