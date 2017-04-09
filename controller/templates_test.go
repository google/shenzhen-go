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

package controller

import (
	"testing"

	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/model/parts"
)

type nopWriter struct{}

func (nopWriter) Write(r []byte) (int, error) { return len(r), nil }

func TestGoTemplate(t *testing.T) {
	// Smoke-testing the template.

	tests := []*model.Graph{
		{
			FilePath:    "filepath",
			URLPath:     "urlpath",
			Name:        "non-command",
			PackagePath: "package/path",
			IsCommand:   false,
		},
		{
			FilePath:    "filepath",
			URLPath:     "urlpath",
			Name:        "command",
			PackagePath: "package/path",
			IsCommand:   true,
		},
		{
			FilePath:    "filepath",
			URLPath:     "urlpath",
			Name:        "has a node and a chanel",
			PackagePath: "package/path",
			IsCommand:   false,
			Nodes: map[string]*model.Node{
				"foo": {
					Part:         &parts.Code{},
					Name:         "foo",
					Enabled:      true,
					Multiplicity: 1,
					Wait:         true,
				},
			},
			Channels: map[string]*model.Channel{
				"bar": {
					Name:      "bar",
					Anonymous: false,
					Type:      "int",
					Capacity:  0,
				},
			},
		},
		{
			FilePath:    "filepath",
			URLPath:     "urlpath",
			Name:        "has a disabled node",
			PackagePath: "package/path",
			IsCommand:   false,
			Nodes: map[string]*model.Node{
				"foo": {
					Part:         &parts.Code{},
					Name:         "foo",
					Enabled:      false,
					Multiplicity: 1,
					Wait:         false,
				},
			},
		},
		{
			FilePath:    "filepath",
			URLPath:     "urlpath",
			Name:        "has a node with multiplicity > 1",
			PackagePath: "package/path",
			IsCommand:   false,
			Nodes: map[string]*model.Node{
				"foo": {
					Part:         &parts.Code{},
					Name:         "foo",
					Enabled:      true,
					Multiplicity: 50,
					Wait:         true,
				},
			},
		},
		{
			FilePath:    "filepath",
			URLPath:     "urlpath",
			Name:        "has a node that isn't waited for",
			PackagePath: "package/path",
			IsCommand:   false,
			Nodes: map[string]*model.Node{
				"foo": {
					Part:         &parts.Code{},
					Name:         "foo",
					Enabled:      true,
					Multiplicity: 1,
					Wait:         false,
				},
			},
		},
	}

	for _, test := range tests {
		if err := goTemplate.Execute(nopWriter{}, test); err != nil {
			t.Errorf("goTemplate.Execute(%v) = %v, want nil error", test.Name, err)
		}
	}
}
