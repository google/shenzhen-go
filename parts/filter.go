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
	"text/template"
)

const filterTemplateSrc = `for x := range {{.Input}} {
    {{range .Paths -}}
    if {{.Pred}} {
        {{.Output}} <- x
    }{{end}}
}
{{- range .Paths}}
close({{.Output}})
{{- end}}`

var filterTemplate = template.Must(template.New("filter").Parse(filterTemplateSrc))

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
	// TODO: Method of adding and removing output paths.
	_, err := tmpl.New("part_view").Parse(`<div class="formfield">
		<label for="FilterInput">Input</label>
		<select name="FilterInput">
			{{range .Graph.Channels -}}
			<option value=".Name" {{if eq .Name $.Node.Part.Input}}selected{{end}}>{{.Name}}</option>
			{{- end}}
		</select>
	</div>
	{{range $index, $path := .Node.Part.Paths}}
	<fieldset>
		<div class="formfield">
			<label for="FilterPath{{$index}}Output">Output</label>
			<select name="FilterPath{{$index}}Output">
				{{range $.Graph.Channels -}}
				<option value=".Name" {{if eq .Name $path.Output}}selected{{end}}>{{.Name}}</option>
				{{- end}}
			</select>
		</div>
		<div class="formfield">
			<label for="FilterPath{{$index}}Predicate">Predicate</label>
			<input type="text" name="FilterPath{{$index}}Predicate" required value="{{$path.Pred}}">
		</div>
	</fieldset>
	{{- end}}`)
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

// Refresh refreshes any cached information.
func (f *Filter) Refresh() error { return nil }

// TypeKey returns "Filter".
func (*Filter) TypeKey() string { return "Filter" }
