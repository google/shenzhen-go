// Copyright 2017 Google Inc.
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
	"html/template"
	"net/http"
	text "text/template"

	"github.com/google/shenzhen-go/v0/source"
)

const (
	aggregatorEditorTemplateSrc = `
<div class="formfield">
	<label for="Input">Input</label>
	<select name="Input">
		{{range .Graph.Channels -}}
		<option value="{{.Name}}" {{if eq .Name $.Node.Part.Input}}selected{{end}}>{{.Name}}</option>
		{{- end}}
	</select>
</div>
<div class="formfield">
	<label for="Output">Output</label>
	<select name="Output">
		{{range .Graph.Channels -}}
		<option value="{{.Name}}" {{if eq .Name $.Node.Part.Output}}selected{{end}}>{{.Name}}</option>
		{{- end}}
	</select>
</div>
<div class="formfield">
	<label for="Aggregation">Aggregation</label>
	<select name="Aggregation">
		<option value="Append" {{if eq "Append" $.Node.Part.Aggregation}}selected{{end}}>Append</option>
		<option value="Sum" {{if eq "Sum" $.Node.Part.Aggregation}}selected{{end}}>Sum</option>
	</select>
</div>
<div class="formfield">
	<label for="Value">Value(x)</label>
	<input type="text" name="Value" required value="{{$.Node.Part.Value}}">
</div>
<div class="formfield">
	<label for="ValueType">Value Type</label>
	<textarea name="ValueType" required>{{$.Node.Part.ValueType}}</textarea>
</div>
<div class="formfield">
	<label for="Key">Key(x)</label>
	<input type="text" name="Key" value="{{$.Node.Part.Key}}">
</div>
<div class="formfield">
	<label for="KeyType">Key Type</label>
	<textarea name="KeyType">{{$.Node.Part.KeyType}}</textarea>
</div>
`

	aggregatorBodyTemplateSrc = `{{if .Key -}}
	agg := make(map[{{.KeyType}}]{{.ValueType}})
	{{else -}}
	var agg {{.ValueType}}
	{{end -}}
	for x := range {{.Input}} {
		{{if .Key -}}
			k, v := {{.Key}}, {{.Value}}
			{{if eq .Aggregation "Append" -}}
			agg[k] = append(agg[k], v)
			{{else if eq .Aggregation "Sum" -}}
			agg[k] += v
			{{end -}}
		{{else -}}
			{{if eq .Aggregation "Append" -}}
			agg = append(agg, {{.Value}})
			{{else if eq .Aggregation "Sum" -}}
			agg += {{.Value}}
			{{end -}}
		{{end -}}
	}
	{{.Output}} <- agg
	`
)

var aggregatorBodyTemplate = text.Must(text.New("aggregator").Parse(aggregatorBodyTemplateSrc))

// Aggregator aggregates some expression of input over some keys.
type Aggregator struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	Aggregation string `json:"aggr"`
	Value       string `json:"value"`
	ValueType   string `json:"value_type"`
	Key         string `json:"key"`
	KeyType     string `json:"key_type"`
}

// AssociateEditor associates a template called "part_view" with the given template.
func (*Aggregator) AssociateEditor(t *template.Template) error {
	_, err := t.New("part_view").Parse(aggregatorEditorTemplateSrc)
	return err
}

// Channels returns any channels used. Anything returned that is not a channel is ignored.
func (p *Aggregator) Channels() (read, written source.StringSet) {
	return source.NewStringSet(p.Input), source.NewStringSet(p.Output)
}

// Clone returns a copy of this part.
func (p *Aggregator) Clone() interface{} {
	s := *p
	return &s
}

// Help returns useful help information.
func (*Aggregator) Help() template.HTML {
	return `<p>
	Aggregator reads data from an input channel. It aggregates (sums, appends, or similar) the result of
	an expression based on each input value <code>x</code>, optionally grouped by a key expression (similarly).
	Once all input has been read, it writes the single result value to the output channel and closes it.
	</p><p>
	<h4>Example 1</h4>
	Suppose Input and Output are both <code>int</code> channels, the Aggregation is "Sum",
    the Value(x) expression is <code>x</code> and Value Type is <code>int</code>. 
	Then the Input sequence
	<p><code>1, 2, 3, 4, 5, (close)</code></p>
	produces the Output sequence
	<p><code>15, (close)</code>.</p>
	</p><p>
	<h4>Example 2</h4>
	Suppose Input is a <code>struct{ K string, V int }</code> channel,
	Output is a <code>map[string]int</code> channel, the Aggregation is "Sum",
    the Value(x) expression is <code>x.V</code>, the Value Type is <code>int</code>, 
	the Key(x) expression is <code>x.K</code>, and the Key Type is <code>string</code>.
	Then the Input sequence
	<p><code>{"odd", 1}, {"even", 2}, {"odd", 3}, {"even", 4}, {"odd", 5}, (close)</code></p>
	produces the Output sequence
	<p><code>map[string]int{"odd": 9, "even": 6}, (close)</code>.</p>
	Note that the Key(x) expression doesn't have to be a field of the input. For similar output to this example,
	consider an input channel of type <code>int</code>, and a Key expression of <code>x % 2</code> 
	(which is type <code>int</code>). 
	</p><p>
	<h4>Example 3 (Loading puppies into a wheelbarrow)</h4>
	Suppose Input is a <code>string</code> channel,
	Output is a <code>map[string][]string</code> channel, the Aggregation is "Append",
    the Value(x) expression is <code>x</code>, the Value Type is <code>[]string</code>,
	the Key(x) expression is <code>"wheelbarrow"</code>, and the Key Type is <code>string</code>.
	Then the Input sequence
	<p><code>"pupper", "doggo", "pupper", "pupper", (close)</code></p>
	produces the Output sequence
	<p><code>map[string][]string{"wheelbarrow": []string{pupper", "doggo", "pupper", "pupper"}}, (close)</code>.</p>
	</p>
	`
}

// Impl returns Go source code implementing the part.
func (p *Aggregator) Impl() (head, body, tail string) {
	buf := new(bytes.Buffer)
	aggregatorBodyTemplate.Execute(buf, p)
	return "", buf.String(), fmt.Sprintf("close(%s)", p.Output)
}

// Imports returns any extra import lines needed.
func (*Aggregator) Imports() []string {
	return nil
}

// RenameChannel renames any uses of the channel "from" to the channel "to".
func (p *Aggregator) RenameChannel(from, to string) {
	if p.Input == from {
		p.Input = to
	}
	if p.Output == from {
		p.Output = to
	}
}

// TypeKey returns the string "Aggregator"
func (*Aggregator) TypeKey() string {
	return "Aggregator"
}

// Update sets fields in the part based on info in the given Request.
func (p *Aggregator) Update(req *http.Request) error {
	if req == nil {
		return nil
	}
	if err := req.ParseForm(); err != nil {
		return err
	}
	p.Input = req.FormValue("Input")
	p.Output = req.FormValue("Output")
	p.Aggregation = req.FormValue("Aggregation")
	p.Value = req.FormValue("Value")
	p.ValueType = req.FormValue("ValueType")
	p.Key = req.FormValue("Key")
	p.KeyType = req.FormValue("KeyType")
	return nil
}
