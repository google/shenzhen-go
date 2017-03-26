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

// Package api holds types needed to communicate with the UI.
package api

// Direction describes which way information flows in a Pin.
type Direction string

// The various directions.
const (
	Input  Direction = "in"
	Output Direction = "out"
)

// Pin represents a connecting pin on a Node.
type Pin struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Direction Direction `json:"dir"`
	Binding   string    `json:"binding,omitempty"` // Channel key it is connected to, or empty.
}

// FullType returns the full pin type, including the <-chan / chan<-.
func (p *Pin) FullType() string {
	c := "<-chan "
	if p.Direction == Output {
		c = "chan<- "
	}
	return c + p.Type
}

// Node represents a goroutine.
type Node struct {
	Name    string          `json:"-"`
	Pins    map[string]*Pin `json:"pins"`
	Enabled bool            `json:"enabled"`

	// Visual position - topleft corner
	X int `json:"x"`
	Y int `json:"y"`
}

// Channel represents connections between pins.
type Channel struct {
	Name     string `json:"-"`
	Type     string `json:"type"`
	Capacity int    `json:"cap"`
}

// Graph represents a package / program / collection of nodes and channels.
type Graph struct {
	Nodes    map[string]*Node    `json:"nodes"`
	Channels map[string]*Channel `json:"channels"`
}
