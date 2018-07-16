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

var unbatchPins = pin.NewMap(
	&pin.Definition{
		Name:      "input",
		Direction: pin.Input,
		Type:      "[]$Any",
	},
	&pin.Definition{
		Name:      "output",
		Direction: pin.Output,
		Type:      "$Any",
	})

func init() {
	model.RegisterPartType("Unbatch", "Flow", &model.PartType{
		New: func() model.Part { return &Unbatch{} },
		Panels: []model.PartPanel{{
			Name: "Help",
			Editor: `<div>
			<p>
				An Unbatch part consumes slices of any type, and sends the 
				individual elements on the output one at a time. After all
				elements are sent and the input is closed, the output will
				be closed.
			</p>
			</div>`,
		}},
	})
}

// Unbatch is a part which consumes slices of any type, and sends the
// individual elements on the output one at a time. After all
// elements are sent and the input is closed, the output will
// be closed.
type Unbatch struct{}

// Clone returns a clone of this Unbatch.
func (Unbatch) Clone() model.Part { return &Unbatch{} }

// Impl returns the Unbatch implementation.
func (Unbatch) Impl(string, bool, map[string]string) model.PartImpl {
	return model.PartImpl{
		Body: `for in := range input { 
		for _, el := range in { 
			output <- el 
		} 
	}`,
		Tail: "close(output)",
	}
}

// Pins returns a map declaring a single input of any slice type
// and a single output of the slice element type.
func (Unbatch) Pins() pin.Map { return unbatchPins }

// TypeKey returns "Unbatch".
func (Unbatch) TypeKey() string { return "Unbatch" }
