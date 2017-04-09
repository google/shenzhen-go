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
	"reflect"
	"strings"
	"testing"

	"github.com/google/shenzhen-go/model/parts"
	"github.com/google/shenzhen-go/model/pin"
)

func TestLoadJSON(t *testing.T) {
	json := strings.NewReader(`{
	"nodes": {
		"foo": {
			"part_type": "Code", 
			"part": {
				"pins": [
					{
						"name": "output",
						"type": "int",
						"dir": "out"
					}
				]
			},
			"connections": {
				"output": "baz",
				"ignored": "ignored"
			}
		},
		"bar": {
			"part_type": "Code", 
			"part": {
				"pins": [
					{
						"name": "input",
						"type": "int",
						"dir": "in"
					}
				]
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
			Part: parts.NewCode(nil, "", "", "", []pin.Definition{
				{
					Name:      "output",
					Direction: pin.Output,
					Type:      "int",
				},
			}),
			Connections: map[string]string{
				"output": "baz",
			},
		},
		"bar": {
			Name:         "bar",
			Multiplicity: 1,
			Part: parts.NewCode(nil, "", "", "", []pin.Definition{
				{
					Name:      "input",
					Direction: pin.Input,
					Type:      "int",
				},
			}),
			Connections: map[string]string{
				"input": "baz",
			},
		},
	}
	if got, want := got.Nodes, wantNodes; !reflect.DeepEqual(got, want) {
		t.Errorf("LoadJSON().Nodes = %#v, want %#v", got, want)
	}
	wantChans := map[string]*Channel{
		"baz": {
			Name: "baz",
			Type: "int",
		},
	}
	if got, want := got.Channels, wantChans; !reflect.DeepEqual(got, want) {
		t.Errorf("LoadJSON().Channels = %#v, want %#v", got, want)
	}
}
