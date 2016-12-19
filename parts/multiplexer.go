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

var multiplexerTmpl = template.Must(template.New("multiplexer").Parse(multiplexerTmplSrc))

// Multiplexer reads from N input channels and writes values into a single output
// channel. All the channels must have the same or compatible types. Once all input
// channels are closed, the output channel is also closed.
type Multiplexer struct {
	Inputs []string
	Output string
}

// Impl returns the content of a goroutine implementing the multiplexer.
func (m *Multiplexer) Impl() string {
	b := new(bytes.Buffer)
	multiplexerTmpl.Execute(b, m)
	return b.String()
}

// ChannelsRead returns the names of all channels read by this goroutine.
func (m *Multiplexer) ChannelsRead() []string { return m.Inputs }

// ChannelsWritten returns the names of all channels written by this goroutine.
func (m *Multiplexer) ChannelsWritten() []string { return []string{m.Output} }
