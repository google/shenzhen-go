// Copyright 2018 Google Inc.
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

package parts

import (
	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

var forEachPins = pin.NewMap(
	&pin.Definition{
		Name:      "input",
		Direction: pin.Input,
		Type:      "$Any",
	},
)

func init() {
	/*
		model.RegisterPartType("ForEach", "General", &model.PartType{
			New: func() model.Part {
				return &ForEach{}
			},
			Panels: []model.PartPanel{
				{
					Name:   "Help",
					Editor: `<div><p>TODO: implement this help</p></div>`,
				},
			},
		})
	*/
}

// ForEach is a BYO-code part that runs something for each input.
// It's like "Code" but the body is put inside a func called per input message.
type ForEach struct {
	Imports []string `json:"imports"`
	Head    []string `json:"head"`
	Body    []string `json:"body"`
	Tail    []string `json:"tail"`
	Outputs pin.Map  `json:"outputs"`
}

// Clone returns a clone of this ForEach part.
func (e *ForEach) Clone() model.Part {
	e0 := *e
	return &e0
}

// Impl returns the implementation for this part.
func (e *ForEach) Impl(string, bool, map[string]string) model.PartImpl {
	return model.PartImpl{}
}

// Pins returns the pin map for this part.
func (e *ForEach) Pins() pin.Map {
	return forEachPins
}

// TypeKey returns "ForEach".
func (e *ForEach) TypeKey() string {
	return "ForEach"
}
