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

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

var transformPins = pin.NewMap(
	&pin.Definition{
		Name:      "inputs",
		Direction: pin.Input,
		Type:      "$AnyIn",
	},
	&pin.Definition{
		Name:      "outputs",
		Direction: pin.Output,
		Type:      "$AnyOut",
	},
)

func init() {
	model.RegisterPartType("Transform", "General", &model.PartType{
		New: func() model.Part { return &Transform{} },
		Panels: []model.PartPanel{
			{
				Name:   "Imports",
				Editor: `<div class="codeedit" id="transform-imports"></div>`,
			},
			{
				Name:   "Transform",
				Editor: `<span class="link" id="transform-format-link">Format</span><div class="codeedit" id="transform-body"></div>`,
			},
			{
				Name: "Help",
				Editor: `<div>
			<p>
				A Transform part converts inputs into outputs. The transformation itself is BYO code.
			</p><p>
				The BYO code body can be any function body but must <code>return</code> the
				transformed input, available as a value called <code>input</code>.
			</p>
			</div>`,
			},
		},
	})
}

// Transform is a part which immediately closes the output channel.
type Transform struct {
	Imports []string `json:"imports"`
	Body    []string `json:"body"`
}

// Clone returns a clone of this Transform.
func (t *Transform) Clone() model.Part {
	return &Transform{Body: t.Body}
}

// Impl returns the Transform implementation.
func (t *Transform) Impl(_ string, _ bool, types map[string]string) model.PartImpl {
	return model.PartImpl{
		Imports: t.Imports,
		Body: fmt.Sprintf(`for input := range inputs {
			outputs <- func() %s {
				%s
			}()
		}`, types["$AnyOut"], strings.Join(t.Body, "\n")),
		Tail: "close(outputs)",
	}
}

// Pins returns a map declaring a single input and single output of any type.
func (t *Transform) Pins() pin.Map { return transformPins }

// TypeKey returns "Transform".
func (Transform) TypeKey() string { return "Transform" }
