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
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strconv"

	"github.com/google/shenzhen-go/dev/source"
)

const (
	broadcastEditorTemplateSrc = `<div class="formfield">
		<label for="BroadcastInput">Input</label>
		<select name="BroadcastInput">
			{{range .Graph.Channels -}}
			<option value="{{.Name}}" {{if eq .Name $.Node.Part.Input}}selected{{end}}>{{.Name}}</option>
			{{- end}}
		</select>
	</div>
	<script>
	var nextFieldset = {{len .Node.Part.Outputs}};
	function anotherFieldsetPlease() {
		nf = document.getElementById('pathtemplate').cloneNode(true);
		nf.id = '';
		var nfc = nf.children;
		for (var i=0; i<nfc.length; i++) {
			var dc = nfc[i].children
			for (var j=0; j<dc.length; j++) {
				var n = dc[j].name;
				if (n) {
					dc[j].name = n + nextFieldset;
					dc[j].htmlRequired = true;
				}
				var f = dc[j].htmlFor;
				if (f) {
					dc[j].htmlFor = f + nextFieldset;
				}
			}
		}
		var ib = document.getElementById('addmoreplz');
		ib.parentNode.insertBefore(nf,ib);
		nextFieldset++;
	}
    function removeMePlease(el) {
		el.parentNode.parentNode.removeChild(el.parentNode);
	}
	</script>
	<fieldset id="pathtemplate">
		<div class="formfield">
			<label for="BroadcastOutput">Output</label>
			<select name="BroadcastOutput">
				{{range $.Graph.Channels -}}
				<option value="{{.Name}}">{{.Name}}</option>
				{{- end}}
			</select>
		</div>
		<a href="javascript:void(0)" onclick="removeMePlease(this)">Remove this output</a>
	</fieldset>
	{{range $index, $path := .Node.Part.Outputs}}
	<fieldset>
		<div class="formfield">
			<label for="BroadcastOutput{{$index}}">Output</label>
			<select name="BroadcastOutput{{$index}}">
				{{range $.Graph.Channels -}}
				<option value="{{.Name}}" {{if eq .Name $path}}selected{{end}}>{{.Name}}</option>
				{{- end}}
			</select>
		</div>
		<a href="javascript:void(0)" onclick="removeMePlease(this)">Remove this output</a>
	</fieldset>
	{{- end }}
	<a id="addmoreplz" href="javascript:void(0)" onclick="anotherFieldsetPlease()">Add output</a>`
)

var broadcastOutputNameRE = regexp.MustCompile(`^BroadcastOutput(\d+)$`)

// Broadcast sends each input message to every output, closing all the outputs
// when the input is closed.
type Broadcast struct {
	Input   string   `json:"input"`
	Outputs []string `json:"outputs"`
}

// AssociateEditor adds a "part_view" template to the given template.
func (b *Broadcast) AssociateEditor(tmpl *template.Template) error {
	_, err := tmpl.New("part_view").Parse(broadcastEditorTemplateSrc)
	return err
}

// Channels returns the names of all channels used by this goroutine.
func (b *Broadcast) Channels() (read, written source.StringSet) {
	return source.NewStringSet(b.Input), source.NewStringSet(b.Outputs...)
}

// Clone returns a copy of this part.
func (b *Broadcast) Clone() interface{} {
	return &Broadcast{
		Input:   b.Input,
		Outputs: append([]string(nil), b.Outputs...),
	}
}

// Help returns useful help information.
func (*Broadcast) Help() template.HTML {
	return `<p>
	Broadcast reads an input channel for values, and sends each value to every output channel. 
	Once the input channel is closed and all values have been sent, every output channel is closed.
	</p>`
}

// Impl returns the content of a goroutine implementation.
func (b *Broadcast) Impl() (head, body, tail string) {
	bod := new(bytes.Buffer)
	fmt.Fprintf(bod, "for x := range %s {\n", b.Input)
	for _, o := range b.Outputs {
		fmt.Fprintf(bod, "\t%s <- x\n", o)
	}
	fmt.Fprintln(bod, "}")

	tl := new(bytes.Buffer)
	for o := range source.NewStringSet(b.Outputs...) {
		fmt.Fprintf(tl, "close(%s)\n", o)
	}
	return "", bod.String(), tl.String()
}

// Imports returns a nil slice.
func (*Broadcast) Imports() []string { return nil }

// RenameChannel renames a channel possibly used by the input or any outputs.
func (b *Broadcast) RenameChannel(from, to string) {
	if b.Input == from {
		b.Input = to
	}
	for i := range b.Outputs {
		if b.Outputs[i] == from {
			b.Outputs[i] = to
		}
	}
}

// TypeKey returns "Broadcast".
func (*Broadcast) TypeKey() string { return "Broadcast" }

// Update sets fields based on the given Request.
func (b *Broadcast) Update(r *http.Request) error {
	if r == nil {
		return nil
	}
	if err := r.ParseForm(); err != nil {
		return err
	}
	p := make(map[int]string)
	var ks []int
	for k := range r.Form {
		if s := broadcastOutputNameRE.FindStringSubmatch(k); s != nil {
			o, err := strconv.Atoi(s[1])
			if err != nil {
				return err
			}
			ks = append(ks, o)
			p[o] = r.FormValue(k)
			continue
		}
	}
	// Ensure the inputs stay in relative order.
	sort.Ints(ks)

	b.Input = r.FormValue("BroadcastInput")
	b.Outputs = nil
	for _, k := range ks {
		v := p[k]
		if v == "" {
			continue
		}
		b.Outputs = append(b.Outputs, v)
	}
	return nil
}
