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
	"strings"
	"testing"

	"gopkg.in/d4l3k/messagediff.v1"

	"github.com/google/shenzhen-go/dev/model/pin"
)

func init() {
	RegisterPartType("Fake", &PartType{
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
			Multiplicity: 1,
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
			Multiplicity: 1,
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
			Type: "int",
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
