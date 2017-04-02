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
)

// Graph represents a package / program / collection of nodes and channels.
type Graph struct {
	SourcePath  string              `json:"-"` // path to the JSON source.
	Name        string              `json:"name"`
	PackagePath string              `json:"package_path"`
	IsCommand   bool                `json:"is_command"`
	Nodes       map[string]*Node    `json:"nodes"`
	Channels    map[string]*Channel `json:"channels"`
}

// NewGraph returns a new empty graph associated with a file path.
func NewGraph(srcPath, pkgPath string) *Graph {
	return &Graph{
		SourcePath:  srcPath,
		PackagePath: pkgPath,
		Channels:    make(map[string]*Channel),
		Nodes:       make(map[string]*Node),
	}
}

// PackageName extracts the name of the package from the package path ("full" package name).
func (g *Graph) PackageName() string {
	i := strings.LastIndex(g.PackagePath, "/")
	if i < 0 {
		return g.PackagePath
	}
	return g.PackagePath[i+1:]
}
