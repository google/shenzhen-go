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
	html "html/template"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"text/template"

	"github.com/google/shenzhen-go/v0/source"
)

const (
	filterBodyTemplateSrc = `for x := range {{.Input}} {
    {{range .Paths}}{{if and .Pred .Output}}
    if {{.Pred}} {
        {{.Output}} <- x
    }{{end}}{{end}}
}`

	filterTailTemplateSrc = `{{range .ChannelsWritten}}
close({{.}})
{{end}}`

	filterEditorTemplateSrc = `<div class="formfield">
		<label for="FilterInput">Input</label>
		<select name="FilterInput">
			{{range .Graph.Channels -}}
			<option value="{{.Name}}" {{if eq .Name $.Node.Part.Input}}selected{{end}}>{{.Name}}</option>
			{{- end}}
		</select>
	</div>
	<script>
	var nextFieldset = {{len .Node.Part.Paths}};
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
			<label for="FilterOutput">Output</label>
			<select name="FilterOutput">
				{{range $.Graph.Channels -}}
				<option value="{{.Name}}">{{.Name}}</option>
				{{- end}}
			</select>
		</div>
		<div class="formfield">
			<label for="FilterPredicate">Predicate(x)</label>
			<input type="text" name="FilterPredicate">
		</div>
		<a href="javascript:void(0)" onclick="removeMePlease(this)">Remove this output</a>
	</fieldset>
	{{range $index, $path := .Node.Part.Paths}}
	<fieldset>
		<div class="formfield">
			<label for="FilterOutput{{$index}}">Output</label>
			<select name="FilterOutput{{$index}}">
				{{range $.Graph.Channels -}}
				<option value="{{.Name}}" {{if eq .Name $path.Output}}selected{{end}}>{{.Name}}</option>
				{{- end}}
			</select>
		</div>
		<div class="formfield">
			<label for="FilterPredicate{{$index}}">Predicate(x)</label>
			<input type="text" name="FilterPredicate{{$index}}" required value="{{$path.Pred}}">
		</div>
		<a href="javascript:void(0)" onclick="removeMePlease(this)">Remove this output</a>
	</fieldset>
	{{- end }}
	<a id="addmoreplz" href="javascript:void(0)" onclick="anotherFieldsetPlease()">Add output</a>`
)

var (
	filterBodyTemplate = template.Must(template.New("filter").Parse(filterBodyTemplateSrc))
	filterTailTemplate = template.Must(template.New("filter").Parse(filterTailTemplateSrc))

	pathOutputNameRE = regexp.MustCompile(`^FilterOutput(\d+)$`)
	pathPredNameRE   = regexp.MustCompile(`^FilterPredicate(\d+)$`)
)

type pathway struct {
	Pred   string `json:"pred"`
	Output string `json:"output"`
}

// Filter tests values from the input and passes it on to one or more
// outputs based on predicates.
type Filter struct {
	Input string    `json:"input"`
	Paths []pathway `json:"paths"`
}

// AssociateEditor adds a "part_view" template to the given template.
func (f *Filter) AssociateEditor(tmpl *html.Template) error {
	_, err := tmpl.New("part_view").Parse(filterEditorTemplateSrc)
	return err
}

// Channels returns the names of all channels used by this goroutine.
func (f *Filter) Channels() (read, written source.StringSet) {
	o := make(source.StringSet, len(f.Paths))
	for _, p := range f.Paths {
		o.Add(p.Output)
	}
	return source.NewStringSet(f.Input), o
}

// ChannelsWritten is a convenience function for the template: allows closing outputs only once.
// (More than one close per channel panics.)
func (f *Filter) ChannelsWritten() []string {
	o := make(source.StringSet, len(f.Paths))
	for _, p := range f.Paths {
		o.Add(p.Output)
	}
	return o.Slice()
}

// Clone returns a copy of this Filter part.
func (f *Filter) Clone() interface{} {
	return &Filter{
		Input: f.Input,
		Paths: append([]pathway(nil), f.Paths...),
	}
}

// Help returns useful help information.
func (*Filter) Help() html.HTML {
	return `<p>
	A Filter reads values from the input channel until it is closed, and 
	tests each value <code>x</code> using any number of <em>predicates in <code>x</code></em> 
	(Go boolean expressions).
	Whenever a predicate evaluates to <code>true</code>, the value <code>x</code> is written to the
	corresponding output channel.
	</p><p>
	For example, if the value <code>2</code> (i.e. <code>x = 2</code>) appears in the input, 
	then all these predicates (which can be used in the Predicate(x) field as-is) will evaluate 
	to <code>true</code> (for different reasons):
	<ul>
	    <li><code>x == 2</code></li>
		<li><code>x != 3</code></li>
		<li><code>x % 2 == 0</code></li>
		<li><code>x &gt; 0 && x &lt; 3</code></li>
        <li><code>true</code></li>
		<li><code>x == x</code></li>
	</ul>
	but these all evaluate to <code>false</code>:
	<ul>
	    <li><code>x != 2</code></li>
		<li><code>x == 3</code></li>
		<li><code>x % 2 == 1</code></li>
		<li><code>x &lt; 0 || x &gt; 3</code></li>
        <li><code>false</code></li>
		<li><code>x != x</code></li>
	</ul>
	</p><p>
	An optional statement can appear before the expression (exactly like a Go <code>if</code> test, since that's how it works). 
	This is	useful for doing type assertions. For example, if the input channel has type <code>interface{}</code>, then
	<code>y, ok := x.(string); ok && y == "banana"</code>
	checks if <code>x</code> contains a string, and matches if it is a string and the string is <code>"banana"</code>.
	Note that even if a type assertion is used, <code>x</code> is sent to the output (not <code>y</code>), and 
	<code>x</code> has the type of the input channel.
	If the input channel has type <code>string</code>, then the type assertion is unnecessary and just <code>x == "banana"</code> works.
	</p><p>
	When the input is closed and every value (that matched a predicate) has been written to its output, all the outputs are closed.
	</p><p>
	An input value doesn't need to match a predicate, and it can match more than one. 
	If it doesn't match, it isn't written to any output.
	If it matches more than one predicate, it is written to all that match.
	</p>
	`
}

// Impl returns the content of a goroutine implementation.
func (f *Filter) Impl() (head, body, tail string) {
	b, t := new(bytes.Buffer), new(bytes.Buffer)
	filterBodyTemplate.Execute(b, f)
	filterTailTemplate.Execute(t, f)
	return "", b.String(), t.String()
}

// Imports returns a nil slice.
func (*Filter) Imports() []string { return nil }

// RenameChannel renames a channel possibly used by the input or any outputs.
func (f *Filter) RenameChannel(from, to string) {
	if f.Input == from {
		f.Input = to
	}
	for i := range f.Paths {
		if f.Paths[i].Output == from {
			f.Paths[i].Output = to
		}
	}
}

// TypeKey returns "Filter".
func (*Filter) TypeKey() string { return "Filter" }

// Update sets fields based on the given Request.
func (f *Filter) Update(r *http.Request) error {
	if r == nil {
		// No secret cached information to refresh.
		return nil
	}
	if err := r.ParseForm(); err != nil {
		return err
	}
	p := make(map[int]pathway)
	var ks []int
	for k := range r.Form {
		if s := pathOutputNameRE.FindStringSubmatch(k); s != nil {
			o, err := strconv.Atoi(s[1])
			if err != nil {
				return err
			}
			ks = append(ks, o)
			p[o] = pathway{Pred: p[o].Pred, Output: r.FormValue(k)}
			continue
		}
		if s := pathPredNameRE.FindStringSubmatch(k); s != nil {
			o, err := strconv.Atoi(s[1])
			if err != nil {
				return err
			}
			p[o] = pathway{Pred: r.FormValue(k), Output: p[o].Output}
		}
	}
	// Ensure the inputs stay in relative order.
	sort.Ints(ks)

	f.Input = r.FormValue("FilterInput")
	f.Paths = nil
	for _, k := range ks {
		v := p[k]
		if v.Output == "" || v.Pred == "" {
			continue
		}
		f.Paths = append(f.Paths, v)
	}
	return nil
}
