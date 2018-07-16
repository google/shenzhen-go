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

var closerPins = pin.NewMap(&pin.Definition{
	Name:      "output",
	Direction: pin.Output,
	Type:      "$Any",
})

func init() {
	model.RegisterPartType("Closer", "Utility", &model.PartType{
		New: func() model.Part { return &Closer{} },
		Panels: []model.PartPanel{{
			Name: "Help",
			Editor: `<div>
			<p>
				A Closer part immediately closes the output channel.
			</p>
			</div>`,
		}},
	})
}

// Closer is a part which immediately closes the output channel.
type Closer struct{}

// Clone returns a clone of this Closer.
func (Closer) Clone() model.Part { return &Closer{} }

// Impl returns the Closer implementation.
func (Closer) Impl(string, bool, map[string]string) model.PartImpl {
	return model.PartImpl{Tail: "close(output)"}
}

// Pins returns a map declaring a single output of any type.
func (Closer) Pins() pin.Map { return closerPins }

// TypeKey returns "Closer".
func (Closer) TypeKey() string { return "Closer" }
