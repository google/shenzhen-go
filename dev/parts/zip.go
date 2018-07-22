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

// Zip implements a "zipper" part, that combines inputs in lockstep.
type Zip struct {
	InputNum   uint          `json:"input_num"`
	FinishMode ZipFinishMode `json:"finish_mode"`
}

func (z Zip) outputType(types map[string]*source.Type) string {
	fs := make([]string, 0, z.InputNum)
	for i := uint(0); i < z.InputNum; i++ {
		tp := fmt.Sprintf("$T%d", i)
		if types == nil {
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
	bb.WriteString("for {\n")
	if z.FinishMode == ZipUntilLastClose {
		bb.WriteString("\tallClosed := true\n")
	}
	for i := uint(0); i < z.InputNum; i++ {
		in := fmt.Sprintf("input%d", i)
		if n.Connections[in] == "nil" {
			continue
		}
		fmt.Fprintf(bb, "\tin%d, open := <- %s\n\tif ", i, in)
		switch z.FinishMode {
		case ZipUntilFirstClose:
			bb.WriteString("!open {\t\tbreak\n\t}\n")
		case ZipUntilLastClose:
			bb.WriteString("open {\t\tallClosed = false\n\t}\n")
		}

		fmt.Fprintf(wb, "\t\tField%d: in%d\n", i, i)
	}
	if z.FinishMode == ZipUntilLastClose {
		bb.WriteString("\tif allClosed { break }\n")
	}
	fmt.Fprintf(bb, "\toutput <- %s{%s}\n}", z.outputType(n.TypeParams), wb.String())
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
