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
	"fmt"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

const keyCounterTypeParam = "$Key"

var keyCounterPins = pin.NewMap(
	&pin.Definition{
		Name:      "input",
		Direction: pin.Input,
		Type:      keyCounterTypeParam,
	},
	&pin.Definition{
		Name:      "output",
		Direction: pin.Output,
		Type:      keyCounterTypeParam,
	},
	&pin.Definition{
		Name:      "result",
		Direction: pin.Output,
		Type:      fmt.Sprintf("map[%s]uint", keyCounterTypeParam),
	})

func init() {
	model.RegisterPartType("KeyCounter", "Utility", &model.PartType{
		New: func() model.Part { return &KeyCounter{} },
		Panels: []model.PartPanel{{
			Name: "Help",
			Editor: `<div>
			<p>
				A KeyCounter produces a frequency table: a map from values to
				how many of them passed through.
			</p><p>
				A KeyCounter passes through values from input to output, counting
				how many of each value it sees in a map. When the input is closed,
				the map is sent on the result channel, and both outputs are closed.
			</p><p>
				Multiplicity affects how many goroutines perform counting, and
				the number of result maps will be equal to the multiplicity.
			</p><p>
				If output is nil (not connected), it is ignored.
			</p>
			</div>`,
		}},
	})
}

// KeyCounter produces a frequency table.
type KeyCounter struct{}

// Clone returns a clone of this Closer.
func (KeyCounter) Clone() model.Part { return &KeyCounter{} }

// Impl returns the Closer implementation.
func (KeyCounter) Impl(types map[string]string) model.PartImpl {
	return model.PartImpl{
		Body: fmt.Sprintf(`
			m := make(map[%s]uint)
			for in := range input { 
				m[in]++
				if output != nil { 
					output <- in 
				} 
			}
			result <- m`, types[keyCounterTypeParam]),
		Tail: `if output != nil { 
			close(output)
		}
		close(result)`,
	}
}

// Pins returns a map declaring an input/output pair of the same type,
// and a result output with map type.
func (KeyCounter) Pins() pin.Map { return keyCounterPins }

// TypeKey returns "Closer".
func (KeyCounter) TypeKey() string { return "KeyCounter" }
