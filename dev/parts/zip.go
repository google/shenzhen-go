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
	"strings"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
	"github.com/google/shenzhen-go/dev/source"
)

// ZipFinishMode is the finish point of a Zip part.
type ZipFinishMode string

// Values for ZipFinishMode.
const (
	ZipUntilFirstClose ZipFinishMode = "first"
	ZipUntilLastClose  ZipFinishMode = "last"
)

func init() {
	model.RegisterPartType("Zip", "Flow", &model.PartType{
		New: func() model.Part {
			return &Zip{
				InputNum:   2,
				FinishMode: ZipUntilFirstClose,
			}
		},
		Panels: []model.PartPanel{
			{
				Name: "Zip",
				Editor: `<div class="form">
				<div class="formfield">
					<label>Number of inputs: <input id="zip-inputnum" type="number"></input></label>
				</div>
				<div class="formfield">
					<label for="zip-finishmode">Finish mode</label>
					<select id="zip-finishmode" name="zip-finishmode">
						<option value="first">Send until first closure</option>
						<option value="last">Send until last closure</option>
					</select>
				</div></div>`,
			},
			{
				Name: "Help",
				Editor: `<div>
			<p>A Zip part combines multiple inputs in lockstep into a single struct.
			The number of inputs is configurable, as is the behaviour on input closure - 
			either to stop sending as soon as any input is closed, or when all inputs
			are closed.</p><p>
			Regardless of finish mode, all input values will be consumed.
			</p>
			</div>`,
			},
		},
	})
}

// Zip implements a "zipper" part, that combines inputs in lockstep.
type Zip struct {
	InputNum   uint          `json:"input_num"`
	FinishMode ZipFinishMode `json:"finish_mode"`
}

func (z Zip) outputType(types map[string]*source.Type) string {
	fs := make([]string, 0, z.InputNum)
	for i := uint(0); i < z.InputNum; i++ {
		tp := fmt.Sprintf("$T%d", i)
		if types != nil {
			tp = types[tp].String()
		}
		fs = append(fs, fmt.Sprintf("Field%d %s", i, tp))

	}
	return "struct { " + strings.Join(fs, ";") + " }"
}

// Clone returns a clone of this part.
func (z Zip) Clone() model.Part { return z }

// Impl returns an implementation for this part.
func (z Zip) Impl(n *model.Node) model.PartImpl {
	bb, wb := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	bb.WriteString(`for {
		allClosed, send := true, true
	`)
	for i := uint(0); i < z.InputNum; i++ {
		input := fmt.Sprintf("input%d", i)
		if n.Connections[input] == "nil" {
			continue
		}
		fmt.Fprintf(bb, `	in%d, open := <- %s
			if open {
				allClosed = false
			}`, i, input)
		if z.FinishMode == ZipUntilFirstClose {
			bb.WriteString(` else {
				send = false
			}
		`)
		}

		fmt.Fprintf(wb, "Field%d: in%d\n", i, i)
	}
	bb.WriteString("if allClosed {\nbreak\n}\n")
	if z.FinishMode == ZipUntilFirstClose {
		bb.WriteString("if send {\n")
	}
	fmt.Fprintf(bb, "output <- %s{%s}\n}", z.outputType(n.TypeParams), wb.String())
	if z.FinishMode == ZipUntilFirstClose {
		bb.WriteString("}\n")
	}

	tail := "close(output)"
	if n.Connections["output"] == "nil" {
		tail = ""
	}
	return model.PartImpl{
		Body: bb.String(),
		Tail: tail,
	}
}

// Pins returns a map with N inputs and 1 output.
func (z Zip) Pins() pin.Map {
	m := pin.NewMap(&pin.Definition{
		Name:      "output",
		Direction: pin.Output,
		Type:      z.outputType(nil),
	})
	for i := uint(0); i < z.InputNum; i++ {
		name := fmt.Sprintf("input%d", i)
		m[name] = &pin.Definition{
			Name:      name,
			Direction: pin.Input,
			Type:      fmt.Sprintf("$T%d", i),
		}
	}
	return m
}

// TypeKey returns "Zip".
func (z Zip) TypeKey() string { return "Zip" }
