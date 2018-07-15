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
	"github.com/google/shenzhen-go/dev/model/pin"
	"github.com/google/shenzhen-go/dev/source"
)

// FakePart fakes a part type for testing purposes.
type FakePart struct {
	Impts []string `json:"imports"`
	Head  string   `json:"head"`
	Body  string   `json:"body"`
	Tail  string   `json:"tail"`
	Pns   pin.Map  `json:"pins"`
}

// Clone returns a shallow copy.
func (f *FakePart) Clone() Part { f2 := *f; return &f2 }

// Imports returns Impts.
func (f *FakePart) Imports() []string { return f.Impts }

// Impl returns Head, Body, Tail.
func (f *FakePart) Impl(map[string]string) (h, b, t string) { return f.Head, f.Body, f.Tail }

// Pins returns Pns.
func (f *FakePart) Pins() pin.Map { return f.Pns }

// TypeKey returns "Fake".
func (f *FakePart) TypeKey() string { return "Fake" }

// TestGraphs contains graphs that are useful for testing.
var TestGraphs = map[string]*Graph{
	"non-command": {
		FilePath:    "filepath",
		URLPath:     "urlpath",
		Name:        "non-command",
		PackagePath: "package/path",
		IsCommand:   false,
	},
	"command": {
		FilePath:    "filepath",
		URLPath:     "urlpath",
		Name:        "command",
		PackagePath: "package/path",
		IsCommand:   true,
	},
	"has a node and a channel": {
		FilePath:    "filepath",
		URLPath:     "urlpath",
		Name:        "has a node and a channel",
		PackagePath: "package/path",
		IsCommand:   false,
		Nodes: map[string]*Node{
			"foo": {
				Part: &FakePart{nil, "", "", "", pin.Map{
					"output": {
						Name:      "output",
						Type:      "int",
						Direction: pin.Output,
					},
				},
				},
				Name:         "foo",
				Enabled:      true,
				Multiplicity: "1",
				Wait:         true,
				Connections: map[string]string{
					"output": "bar",
				},
			},
		},
		Channels: map[string]*Channel{
			"bar": {
				Name:     "bar",
				Type:     source.MustNewType("", "int"),
				Capacity: 0,
			},
		},
	},
	"has a disabled node": {
		FilePath:    "filepath",
		URLPath:     "urlpath",
		Name:        "has a disabled node",
		PackagePath: "package/path",
		IsCommand:   false,
		Nodes: map[string]*Node{
			"foo": {
				Part: &FakePart{nil, "", "", "", pin.Map{
					"output": {
						Name:      "output",
						Type:      "int",
						Direction: pin.Output,
					},
				},
				},
				Name:         "foo",
				Enabled:      false,
				Multiplicity: "1",
				Wait:         false,
				Connections: map[string]string{
					"output": "nil",
				},
			},
		},
	},
	"has a node with multiplicity > 1": {
		FilePath:    "filepath",
		URLPath:     "urlpath",
		Name:        "has a node with multiplicity > 1",
		PackagePath: "package/path",
		IsCommand:   false,
		Nodes: map[string]*Node{
			"foo": {
				Part: &FakePart{nil, "", "", "", pin.Map{
					"output": {
						Name:      "output",
						Type:      "int",
						Direction: pin.Output,
					},
				}},
				Name:         "foo",
				Enabled:      true,
				Multiplicity: "50",
				Wait:         true,
				Connections: map[string]string{
					"output": "nil",
				},
			},
		},
	},
	"has a node that isn't waited for": {
		FilePath:    "filepath",
		URLPath:     "urlpath",
		Name:        "has a node that isn't waited for",
		PackagePath: "package/path",
		IsCommand:   false,
		Nodes: map[string]*Node{
			"foo": {
				Part: &FakePart{nil, "", "", "", pin.Map{
					"output": {
						Name:      "output",
						Type:      "int",
						Direction: pin.Output,
					},
				}},
				Name:         "foo",
				Enabled:      true,
				Multiplicity: "1",
				Wait:         false,
				Connections: map[string]string{
					"output": "nil",
				},
			},
		},
	},
}
