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
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/google/shenzhen-go/dev/model/pin"
)

// Part abstracts the implementation of a node. Concrete implementations should be
// able to be marshalled to and unmarshalled from JSON sensibly.
type Part interface {
	// Clone returns a copy of this part.
	Clone() Part

	// TODO: Aspects supported by the part: multiplicity? variadic? generic?

	// Impl returns Go source code implementing the part.
	// The head is executed, then the body is executed (# Multiplicity
	// instances of the body concurrently), then the tail (once the body/bodies
	// are finished).
	//
	// This allows cleanly closing channels for nodes with Multiplicity > 1.
	// The tail is deferred so that the body can use "return" and it is still
	// executed.
	//
	// The types map indicates inferred types which the part is responsible
	// for interpolating into the output as needed.
	Impl(types map[string]string) (head, body, tail string)

	// Imports returns any extra import lines needed for the Part.
	Imports() []string

	// Pins returns any pins - "channel arguments" - to the part.
	// inputs and outputs map argument names to types (the "<-chan" /
	// "chan<-" part of the type is implied). The map must be keyed
	// by name.
	Pins() pin.Map

	// TypeKey returns the "type" of part.
	TypeKey() string
}

// PartType has metadata common to this type of part, and is also a part factory.
// The HTML is loaded with the editor.
type PartType struct {
	New    func() Part
	Panels []PartPanel
}

// PartPanel describes one panel of the editor interface specific to a part type.
type PartPanel struct {
	Name   string
	Editor template.HTML
}

// PartTypes translates part type strings into useful information.
var PartTypes = make(map[string]*PartType)

// RegisterPartType adds a part type to the PartTypes map.
// This should be used by part types during init.
func RegisterPartType(name string, pt *PartType) {
	PartTypes[name] = pt
}

// PartJSON is a convenient JSON-plus-type-key type.
type PartJSON struct {
	Part json.RawMessage `json:"part,omitempty"`
	Type string          `json:"part_type,omitempty"`
}

// MarshalPart turns a rich Part into JSON-with-type.
func MarshalPart(p Part) (*PartJSON, error) {
	m, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return &PartJSON{Part: m, Type: p.TypeKey()}, nil
}

// Unmarshal converts the JSON into a Part, via the type key.
func (pj *PartJSON) Unmarshal() (Part, error) {
	pt, ok := PartTypes[pj.Type]
	if !ok {
		return nil, fmt.Errorf("unknown part type %q", pj.Type)
	}
	p := pt.New()
	if err := json.Unmarshal(pj.Part, p); err != nil {
		return nil, err
	}
	return p, nil
}
