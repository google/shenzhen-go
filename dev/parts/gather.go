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

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

func init() {
	model.RegisterPartType("Gather", "Flow", &model.PartType{
		New: func() model.Part { return &Gather{InputNum: 2} },
		Panels: []model.PartPanel{
			{
				Name: "Gather",
				Editor: `<div class="form"><div class="formfield">
					<label>Number of inputs: <input id="gather-inputnum" type="number"></input></label>
				</div></div>`,
			},
			{
				Name: "Help",
				Editor: `<div>
			<p>
				A Gather part sends every value received from every input to the output.
				The number of inputs is configurable.
			</p><p>
				Gather is useful for combining multiple outputs. While a single channel
				can be attached to multiple outputs, it can cause a panic if both outputs
				try to close the channel.
			</p>
			</div>`,
			},
		},
	})
}

// Gather is a part type which reads a configurable number of inputs
// and sends values to a single output.
type Gather struct {
	InputNum uint `json:"input_num"`
}

// Clone returns a clone of this part.
func (g Gather) Clone() model.Part { return g }

// Impl returns a "straightforward" for-select implementation.
// Compared with the N-goroutine approach, this doesn't require a WaitGroup
// and has less hidden-buffer (won't read from inputs if blocked on output).
func (g Gather) Impl(n *model.Node) model.PartImpl {
	lb, sb := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	lb.WriteString(`for {
		if true `)
	sb.WriteString("select {\n")
	for i := uint(0); i < g.InputNum; i++ {
		name := fmt.Sprintf("input%d", i)
		if n.Connections[name] == "nil" {
			continue
		}
		fmt.Fprintf(lb, " && %s == nil", name)
		fmt.Fprintf(sb, `case in, open := <- %s:
			if !open { %s = nil; break }
			output <- in
			`, name, name)
	}
	lb.WriteString("{ break }\n")
	sb.WriteString("}\n")
	lb.Write(sb.Bytes())
	lb.WriteString("}\n")
	return model.PartImpl{
		Body: lb.String(),
		Tail: `close(output)`,
	}
}

// Pins returns a map with N inputs and 1 output.
func (g Gather) Pins() pin.Map {
	m := pin.NewMap(&pin.Definition{
		Name:      "output",
		Direction: pin.Output,
		Type:      "$Any",
	})
	for i := uint(0); i < g.InputNum; i++ {
		n := fmt.Sprintf("input%d", i)
		m[n] = &pin.Definition{
			Name:      n,
			Direction: pin.Input,
			Type:      "$Any",
		}
	}
	return m
}

// TypeKey returns "Gather".
func (g Gather) TypeKey() string { return "Gather" }
