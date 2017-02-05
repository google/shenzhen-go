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

	"github.com/google/shenzhen-go/source"
)

const httpServerEditorTemplateSrc = `
<div class="formfield">
	<label for="Address">Listen-address input</label>
	<select name="Address">
		{{range .Graph.Channels -}}
		<option value="{{.Name}}" {{if eq .Name $.Node.Part.Address}}selected{{end}}>{{.Name}}</option>
		{{- end}}
	</select>
</div>
<div class="formfield">
	<label for="Errors">Errors output</label>
	<select name="Errors">
		{{range .Graph.Channels -}}
		<option value="{{.Name}}" {{if eq .Name $.Node.Part.Errors}}selected{{end}}>{{.Name}}</option>
		{{- end}}
	</select>
</div>
<script>
	var nextFieldset = {{len .Node.Part.Handlers}};
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
			<label for="HandlerPattern">Pattern</label>
			<input type="text" name="HandlerPattern">
		</div>
		<div class="formfield">
			<label for="HandlerOutput">Output</label>
			<select name="HandlerOutput">
				{{range $.Graph.Channels -}}
				<option value="{{.Name}}">{{.Name}}</option>
				{{- end}}
			</select>
		</div>
		<a href="javascript:void(0)" onclick="removeMePlease(this)">Remove this handler</a>
	</fieldset>
	{{range $index, $path := .Node.Part.Paths}}
	<fieldset>
    	<div class="formfield">
			<label for="HandlerPattern{{$index}}">Pattern</label>
			<input type="text" name="HandlerPattern{{$index}}" required value="{{$path.Pattern}}">
		</div>
		<div class="formfield">
			<label for="HandlerOutput{{$index}}">Output</label>
			<select name="HandlerOutput{{$index}}">
				{{range $.Graph.Channels -}}
				<option value="{{.Name}}" {{if eq .Name $path.Channel}}selected{{end}}>{{.Name}}</option>
				{{- end}}
			</select>
		</div>
		<a href="javascript:void(0)" onclick="removeMePlease(this)">Remove this handler</a>
	</fieldset>
	{{- end }}
	<a id="addmoreplz" href="javascript:void(0)" onclick="anotherFieldsetPlease()">Add handler</a>`

var httpServerBodyTemplate = text.Must(text.New("httpserver").Parse(`srv := http.Server{
	Addr:    <-{{.Address}},
	Handler: mux,
}
{{.Errors}} <- srv.ListenAndServe()`))

// HTTPServer is a simple HTTP server that farms out requests to channels of type partlib.HTTPRequest,
// as defined by the handlers map.
type HTTPServer struct {
	Address  string            `json:"address"`
	Handlers map[string]string `json:"handlers"`
	Errors   string            `json:"errors"`
}

// AssociateEditor associates a template called "part_view" with the given template.
func (*HTTPServer) AssociateEditor(t *template.Template) error {
	_, err := t.New("part_view").Parse(httpServerEditorTemplateSrc)
	return err
}

// Channels returns any channels used. Anything returned that is not a channel is ignored.
func (p *HTTPServer) Channels() (read, written source.StringSet) {
	return source.NewStringSet(p.Address), p.chansWr()
}

func (p *HTTPServer) chansWr() source.StringSet {
	wr := make(source.StringSet, len(p.Handlers))
	wr.Add(p.Errors)
	for _, h := range p.Handlers {
		wr.Add(h)
	}
	return wr
}

// Clone returns a copy of this part.
func (p *HTTPServer) Clone() interface{} {
	s := *p
	return &s
}

// Help returns useful help information.
func (*HTTPServer) Help() template.HTML {
	return `<blink><h1>TODO</h1></blink>` // TODO: Return helpful information here
}

// Impl returns Go source code implementing the part.
func (p *HTTPServer) Impl() (head, body, tail string) {
	h := new(bytes.Buffer)
	fmt.Fprintf(h, "mux := http.NewServeMux()\n")
	for pat, ch := range p.Handlers {
		fmt.Fprintf(h, "mux.Handle(%q, partlib.HTTPHandlerChan(%s))\n", pat, ch)
	}
	b := new(bytes.Buffer)
	httpServerBodyTemplate.Execute(b, p)
	t := new(bytes.Buffer)
	for ch := range p.chansWr() {
		fmt.Fprintf(t, "close(%s)\n", ch)
	}
	return h.String(), b.String(), t.String()
}

// Imports returns any extra import lines needed.
func (*HTTPServer) Imports() []string {
	return []string{
		`"net/http"`,
		`"github.com/google/shenzhen-go/parts/partlib"`,
	}
}

// Paths returns each handler as a struct. This is for the benefit of the editor template.
func (p *HTTPServer) Paths() []struct{ Pattern, Channel string } {
	r := make([]struct{ Pattern, Channel string }, 0, len(p.Handlers))
	for pat, ch := range p.Handlers {
		r = append(r, struct{ Pattern, Channel string }{pat, ch})
	}
	return r
}

// RenameChannel renames any uses of the channel "from" to the channel "to".
func (p *HTTPServer) RenameChannel(from, to string) {
	if p.Address == from {
		p.Address = to
	}
	if p.Errors == from {
		p.Errors = to
	}
	for pat, ch := range p.Handlers {
		if ch == from {
			p.Handlers[pat] = to
		}
	}
}

// TypeKey returns the string "HTTPServer"
func (*HTTPServer) TypeKey() string {
	return "HTTPServer"
}

// Update sets fields in the part based on info in the given Request.
func (p *HTTPServer) Update(req *http.Request) error {
	if req == nil {
		return nil
	}
	if err := req.ParseForm(); err != nil {
		return err
	}
	p.Address = req.FormValue("Address")
	p.Errors = req.FormValue("Errors")
	p.Handlers = make(map[string]string)
	for i := 0; ; i++ {
		pat := req.FormValue(fmt.Sprintf("HandlerPattern%d", i))
		ch := req.FormValue(fmt.Sprintf("HandlerOutput%d", i))
		if pat == "" || ch == "" {
			break
		}
		p.Handlers[pat] = ch
	}
	return nil
}
