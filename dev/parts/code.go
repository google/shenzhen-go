// Copyright 2016 Google Inc.
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
	"encoding/json"
	"go/format"
	"log"
	"strings"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

// CodePanels are subpanels for editing code-type parts.
var CodePanels = []model.PartPanel{
	{
		Name:   "Pins",
		Editor: `<div class="codeedit" id="code-pins"></div>`,
	},
	{
		Name:   "Imports",
		Editor: `<div class="codeedit" id="code-imports"></div>`,
	},
	{
		Name:   "Head",
		Editor: `<div class="codeedit" id="code-head"></div>`,
	},
	{
		Name:   "Body",
		Editor: `<div class="codeedit" id="code-body"></div>`,
	},
	{
		Name:   "Tail",
		Editor: `<div class="codeedit" id="code-tail"></div>`,
	},
	{
		Name: "Help",
		Editor: `<div>
	<p>
		A Code part runs (executes) any Go code that you write. It is completely customisable.
	</p><p>
		The first two configuration panels are Pins and Imports. Pins configures the pins available,
		and Imports configures the contents of the imports declaration needed for any code.
	</p><p>
		The actual Go code configuration consists of 3 parts: a Head, a Body, and a Tail.
		First Head is run, then some number of instances of the Body, then the Tail.
	</p><p>
		Body is where the bulk of the code is usually written, as the Multiplicity parameter
		causes multiple copies of the Body to be run. Multiplicity has no effect on Head or Tail.
		The Head and Tail are useful for operations that should only be done once. For example, any 
		output channels written to in the Body can be correctly closed (if desired) in the Tail.
	</p><p>
		Each instance of the Body can use the <code>int</code> parameters <code>instanceNumber</code>
		and <code>multiplicity</code> to distinguish which instance is running and how many are running, 
		if necessary. These parameters always satisfy the relation 
		<code>0 <= instanceNumber < multiplicity</code>.
	</p><p>
		The <code>return</code> statement is allowed but optional in Code. There are no values that
		need to be returned.
		Using <code>return</code> in the Head will prevent the Body or Tail from executing, but 
		using <code>return</code> in the Body won't affect whether the Tail is executed.
	</p>
	</div>
	`,
	},
}

func init() {
	model.RegisterPartType("Code", &model.PartType{
		New: func() model.Part {
			return &Code{pins: pin.NewMap()}
		},
		Panels: CodePanels,
	})
}

// Code is a component containing arbitrary code.
type Code struct {
	imports          []string
	head, body, tail string
	pins             pin.Map
}

// NewCode just makes a new *Code.
func NewCode(imports []string, head, body, tail string, pins pin.Map) *Code {
	return &Code{
		imports: imports,
		head:    head,
		body:    body,
		tail:    tail,
		pins:    pins,
	}
}

type jsonCode struct {
	Imports []string `json:"imports"`
	Head    []string `json:"head"`
	Body    []string `json:"body"`
	Tail    []string `json:"tail"`
	Pins    pin.Map  `json:"pins"`
}

// MarshalJSON encodes the Code component as JSON.
func (c *Code) MarshalJSON() ([]byte, error) {
	k := &jsonCode{
		Imports: c.imports,
		Head:    strings.Split(c.head, "\n"),
		Body:    strings.Split(c.body, "\n"),
		Tail:    strings.Split(c.tail, "\n"),
		Pins:    c.pins,
	}
	stripCR(k.Imports)
	stripCR(k.Head)
	stripCR(k.Body)
	stripCR(k.Tail)
	return json.Marshal(k)
}

// UnmarshalJSON decodes the Code component from JSON.
func (c *Code) UnmarshalJSON(j []byte) error {
	var mp jsonCode
	if err := json.Unmarshal(j, &mp); err != nil {
		return err
	}
	h := strings.Join(mp.Head, "\n")
	b := strings.Join(mp.Body, "\n")
	t := strings.Join(mp.Tail, "\n")
	if err := c.refresh(h, b, t); err != nil {
		// TODO: revisit all this
		log.Printf("Couldn't format or determine channels used: %v", err)
	}
	c.imports = mp.Imports
	c.pins = mp.Pins
	return nil
}

// Pins returns pins. These are 100% user-defined.
func (c *Code) Pins() pin.Map { return c.pins }

// Clone returns a copy of this Code part.
func (c *Code) Clone() model.Part {
	pins := make(pin.Map, len(c.pins))
	for k, v := range c.pins {
		p := *v
		pins[k] = &p
	}
	return &Code{
		imports: c.imports,
		head:    c.head,
		body:    c.body,
		tail:    c.tail,
		pins:    pins,
	}
}

// Impl returns the implementation of the goroutine.
func (c *Code) Impl(map[string]string) (Head, Body, Tail string) {
	// TODO(josh): Figure out the least awful way of letting the
	// user use the types map.
	return c.head, c.body, c.tail
}

// Imports returns a nil slice.
func (c *Code) Imports() []string { return c.imports }

// TypeKey returns "Code".
func (*Code) TypeKey() string { return "Code" }

func (c *Code) refresh(h, b, t string) error {
	// At least save what the user entered.
	c.head, c.body, c.tail = h, b, t

	// Try to format it.
	hf, err := format.Source([]byte(h))
	if err != nil {
		return err
	}
	bf, err := format.Source([]byte(b))
	if err != nil {
		return err
	}
	tf, err := format.Source([]byte(t))
	if err != nil {
		return err
	}

	c.head, c.body, c.tail = string(hf), string(bf), string(tf)
	return nil
}
