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
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"text/template"
)

const (
	filterTemplateSrc = `for x := range {{.Input}} {
    {{range .Paths}}{{if and .Pred .Output}}
    if {{.Pred}} {
        {{.Output}} <- x
    }{{end}}{{end}}
}
{{- range .Paths}}
close({{.Output}})
{{- end}}`

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
			<label for="FilterPredicate">Predicate</label>
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
			<label for="FilterPredicate{{$index}}">Predicate</label>
			<input type="text" name="FilterPredicate{{$index}}" required value="{{$path.Pred}}">
		</div>
		<a href="javascript:void(0)" onclick="removeMePlease(this)">Remove this output</a>
	</fieldset>
	{{- end }}
	<a id="addmoreplz" href="javascript:void(0)" onclick="anotherFieldsetPlease()">Add output</a>`
)

var (
	filterTemplate = template.Must(template.New("filter").Parse(filterTemplateSrc))

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
func (f *Filter) Channels() (read, written []string) {
	o := make([]string, 0, len(f.Paths))
	for _, p := range f.Paths {
		o = append(o, p.Output)
	}
	return []string{f.Input}, o
}

// Impl returns the content of a goroutine implementation.
func (f *Filter) Impl() string {
	b := new(bytes.Buffer)
	filterTemplate.Execute(b, f)
	return b.String()
}

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

	log.Print(p)
	log.Print(ks)

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

// TypeKey returns "Filter".
func (*Filter) TypeKey() string { return "Filter" }
