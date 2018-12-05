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
	"strings"

	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/model/pin"
)

func init() {
	model.RegisterPartType("Transform", "General", &model.PartType{
		New: func() model.Part {
			return &Transform{
				InputType:  "$AnyIn",
				OutputType: "$AnyOut",
			}
		},
		Panels: []model.PartPanel{
			{
				Name: "Types",
				Editor: `<div class="form">
					<div class="formfield">
						<label for="transform-inputtype">Input type</label>
						<input id="transform-inputtype" name="transform-inputtype" type="text"></input>
					</div>
					<div class="formfield">
						<label for="transform-outputtype">Output type</label>
						<input id="transform-outputtype" name="transform-outputtype" type="text"></input>
					</div>
				</div>`,
			},
			{
				Name:   "Imports",
				Editor: `<div class="codeedit" id="transform-imports"></div>`,
			},
			{
				Name: "Transform",
				Editor: `<div class="formfield">
					<span class="link" id="transform-format-link">Format</span>
				</div>
				<div class="codeedit formfield" id="transform-body"></div>`,
			},
			{
				Name: "Help",
				Editor: `<div>
			<p>
				A Transform part converts inputs into outputs. The transformation itself is BYO code.
			</p><p>
				The BYO code body can be any function body but must transform or filter the
				input value (available as a value called <code>input</code>) and write any output
				to <code>outputs</code>.
			</p>
			</div>`,
			},
		},
	})
}

// Transform is a part which immediately closes the output channel.
type Transform struct {
	Imports    []string `json:"imports"`
	Body       []string `json:"body"`
	InputType  string   `json:"input_type"`
	OutputType string   `json:"output_type"`
}

// Clone returns a clone of this Transform.
func (t *Transform) Clone() model.Part {
	return &Transform{Body: t.Body}
}

// Impl returns the Transform implementation.
func (t *Transform) Impl(n *model.Node) model.PartImpl {
	return model.PartImpl{
		Imports: t.Imports,
		Body: fmt.Sprintf(`for input := range inputs {
			func() {
				%s
			}()
		}`, strings.Join(t.Body, "\n")),
		Tail: "if outputs != nil { close(outputs) }",
	}
}

// Pins returns a map declaring a single input and single output of any type.
func (t *Transform) Pins() pin.Map {
	return pin.NewMap(
		&pin.Definition{
			Name:      "inputs",
			Direction: pin.Input,
			Type:      t.InputType,
		},
		&pin.Definition{
			Name:      "outputs",
			Direction: pin.Output,
			Type:      t.OutputType,
		},
	)
}

// TypeKey returns "Transform".
func (Transform) TypeKey() string { return "Transform" }
