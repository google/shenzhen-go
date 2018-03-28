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

// Package pin has types for describing pins (connection points).
package pin

// Direction describes which way information flows in a Pin.
type Direction string

// The various directions.
const (
	Input  Direction = "in"
	Output Direction = "out"
)

// Type returns either "<-chan" or "chan<-" (input or output).
func (d *Direction) Type() string {
	switch *d {
	case Input:
		return "<-chan"
	case Output:
		return "chan<-"
	}
	return ""
}

// Definition describes the main properties of a pin.
type Definition struct {
	Name      string    `json:"-"`
	Type      string    `json:"type"`
	Direction Direction `json:"dir"`
}

// FullType returns the full pin type, including the <-chan / chan<-.
func (d *Definition) FullType() string {
	return d.Direction.Type() + " " + d.Type
}

// Map is a map from pin names to pin definitions.
type Map map[string]*Definition

// FillNames copies map keys into name fields.
func (m Map) FillNames() {
	for n, p := range m {
		p.Name = n
	}
}
