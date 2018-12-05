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
	"bytes"
	"fmt"

	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/model/pin"
)

func init() {
	model.RegisterPartType("Broadcast", "Flow", &model.PartType{
		New: func() model.Part { return &Broadcast{OutputNum: 2} },
		Panels: []model.PartPanel{
			{
				Name: "Broadcast",
				Editor: `<div class="form"><div class="formfield">
					<label>Number of outputs: <input id="broadcast-outputnum" type="number"></input></label>
				</div></div>`,
			},
			{
				Name: "Help",
				Editor: `<div>
			<p>A Broadcast part copies every input value to all of the outputs. 
			The number of outputs is configurable.</p>
			</div>`,
			},
		},
	})
}

// Broadcast is a part that repeats a copy of each input messge to a
// configurable number of ouptuts.
type Broadcast struct {
	OutputNum uint `json:"output_num"`
}

// Clone returns a clone of this part.
func (b Broadcast) Clone() model.Part { return b }

// Impl returns the implementation.
func (b Broadcast) Impl(n *model.Node) model.PartImpl {
	bb, tb := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	bb.WriteString("for in := range input {\n")
	for i := uint(0); i < b.OutputNum; i++ {
		o := fmt.Sprintf("output%d", i)
		if n.Connections[o] == "nil" {
			// We know at design time whether a pin is nil.
			continue
		}
		fmt.Fprintf(bb, "\t%s <- in\n", o)
		fmt.Fprintf(tb, "close(%s)\n", o)
	}
	bb.WriteString("}")
	return model.PartImpl{
		Body: bb.String(),
		Tail: tb.String(),
	}
}

// Pins returns a map with one input and N outputs.
func (b Broadcast) Pins() pin.Map {
	m := pin.NewMap(&pin.Definition{
		Name:      "input",
		Direction: pin.Input,
		Type:      "$Any",
	})
	for i := uint(0); i < b.OutputNum; i++ {
		n := fmt.Sprintf("output%d", i)
		m[n] = &pin.Definition{
			Name:      n,
			Direction: pin.Output,
			Type:      "$Any",
		}
	}
	return m
}

// TypeKey returns "Broadcast".
func (b Broadcast) TypeKey() string { return "Broadcast" }
