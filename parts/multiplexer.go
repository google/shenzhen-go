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
	"bytes"
	"text/template"

	"shenzhen-go/graph"
)

const multiplexerTmplSrc = `var mxwg sync.WaitGroup
{{range $.Inputs}}
mxwg.Add(1)
go func() {
    for x := range {{.}} {
        {{$.Output}} <- x
    }
    mxwg.Done()
}()
{{end}}
mxwg.Wait()
close({{$.Output}})
`

var (
	multiplexerTmpl = template.Must(template.New("multiplexer").Parse(multiplexerTmplSrc))

	// While being developed, check the interface is matched.
	_ = graph.Part(&Multiplexer{})
)

// Multiplexer reads from N input channels and writes values into a single output
// channel. All the channels must have the same or compatible types. Once all input
// channels are closed, the output channel is also closed.
type Multiplexer struct {
	Inputs []string
	Output string
}

// Channels returns the names of all channels used by this goroutine.
func (m *Multiplexer) Channels() (read, written []string) { return m.Inputs, []string{m.Output} }

// Impl returns the content of a goroutine implementing the multiplexer.
func (m *Multiplexer) Impl() string {
	b := new(bytes.Buffer)
	multiplexerTmpl.Execute(b, m)
	return b.String()
}

// Refresh refreshes any cached information.
func (m *Multiplexer) Refresh(g *graph.Graph) error {
	return nil
}
