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

import (
	"encoding/json"

	"github.com/google/shenzhen-go/dev/source"
)

// Direction describes which way information flows in a Pin.
type Direction string

// The various directions.
const (
	Input  Direction = "in"
	Output Direction = "out"
)

// Type returns either "<-chan" or "chan<-" (input or output).
func (d Direction) Type() string {
	switch d {
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

	typePattern *source.TypePattern // lazy-loaded
}

// TypePattern returns a TypePattern for the Type of this pin definition.
func (d *Definition) TypePattern() *source.TypePattern {
	if d.typePattern != nil && d.typePattern.String() == d.Type {
		return d.typePattern
	}
	d.typePattern = source.NewTypePattern(d.Type)
	return d.typePattern
}

// Map is a map from pin names to pin definitions.
type Map map[string]*Definition

// UnmarshalJSON unmarshals the map the usual way, and then
// calls FillNames.
func (m *Map) UnmarshalJSON(b []byte) error {
	var m0 map[string]*Definition
	if err := json.Unmarshal(b, &m0); err != nil {
		return err
	}
	*m = m0
	for n, p := range m0 {
		p.Name = n
	}
	return nil
}
