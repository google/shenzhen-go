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
	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/model/pin"
)

var sinkPins = pin.NewMap(&pin.Definition{
	Name:      "input",
	Direction: pin.Input,
	Type:      "$Any",
})

func init() {
	model.RegisterPartType("Sink", "Utility", &model.PartType{
		New: func() model.Part { return &Sink{} },
		Panels: []model.PartPanel{{
			Name: "Help",
			Editor: `<div>
			<p>
				A Sink part consumes all input and does nothing with it. 
				It completes when the input is exhausted and closed.
			</p>
			</div>`,
		}},
	})
}

// Sink is a part which consumes all input and does nothing with it.
// It completes when the input closes.
type Sink struct{}

// Clone returns a clone of this Sink.
func (Sink) Clone() model.Part { return &Sink{} }

// Impl returns the Sink implementation.
func (Sink) Impl(*model.Node) model.PartImpl {
	return model.PartImpl{Body: "for range input {}"}
}

// Pins returns a map declaring a single input of any type.
func (Sink) Pins() pin.Map { return sinkPins }

// TypeKey returns "Sink".
func (Sink) TypeKey() string { return "Sink" }
