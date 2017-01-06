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

package view

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/shenzhen-go/graph"
	"github.com/google/shenzhen-go/parts"
)

// TODO: Replace these cobbled-together UIs with Polymer or something.
const nodeEditorTemplateSrc = `{{with .Node}}
<head>
	<title>{{if .Name}}{{.Name}}{{else}}[New]{{end}}</title><style>` + css + `</style>
</head>
<body>
	<h1>{{if .Name}}{{.Name}}{{else}}[New]{{end}}</h1>
	<form method="post">
		<div class="formfield">
			<label for="Name">Name</label>
			<input name="Name" type="text" required value="{{.Name}}">
		</div>
		<div class="formfield">
			<label for="Multiplicity">Multiplicity</label>
			<input name="Multiplicity" type="text" required pattern="^[1-9][0-9]*$" title="Must be a whole number, at least 1." value="{{.Multiplicity}}">
		</div>
		<div class="formfield">
			<label for="Wait">Wait for this to finish</label>
			<input name="Wait" type="checkbox" {{if .Wait}}checked{{end}}>
		</div>
		<div class="formfield">
			{{block "part_view" .Part -}}
			<textarea name="Code" rows="25" cols="80">{{.Impl}}</textarea>
			{{- end}}
		</div>
		<div class="formfield hcentre">
			<input type="submit" value="Save">
			<input type="button" value="Return" onclick="window.location.href='?'">
		</div>
	</form>
</body>
{{end}}`

var nodeEditorTemplate = template.Must(template.New("nodeEditor").Parse(nodeEditorTemplateSrc))

func renderNodeEditor(dst io.Writer, g *graph.Graph, n *graph.Node) error {
	return nodeEditorTemplate.Execute(dst, struct {
		Graph *graph.Graph
		Node  *graph.Node
	}{g, n})
}

// Node handles viewing/editing a node.
func Node(g *graph.Graph, name string, w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	n, found := g.Nodes[name]
	if name != "new" && !found {
		http.Error(w, fmt.Sprintf("Node %q not found", name), http.StatusNotFound)
		return
	}
	if n == nil {
		n = &graph.Node{
			Part: &parts.Code{},
		}
	}

	switch r.Method {
	case "POST":
		if err := r.ParseForm(); err != nil {
			log.Printf("Could not parse form: %v", err)
			http.Error(w, "Could not parse", http.StatusBadRequest)
			return
		}

		nm := strings.TrimSpace(r.FormValue("Name"))
		if nm == "" {
			log.Printf("Name invalid [%q == \"\"]", nm)
			http.Error(w, "Name invalid", http.StatusBadRequest)
			return
		}

		mult, err := strconv.Atoi(r.FormValue("Multiplicity"))
		if err != nil {
			log.Printf("Multiplicity is not an integer: %v", err)
			http.Error(w, "Multiplicity is not an integer", http.StatusBadRequest)
			return
		}
		if mult < 1 {
			log.Printf("Must specify positive Multiplicity [%d < 1]", mult)
			http.Error(w, "Multiplicity must be positive", http.StatusBadRequest)
			return
		}

		n.Multiplicity = uint(mult)
		n.Wait = (r.FormValue("Wait") == "on")
		if p, ok := n.Part.(*parts.Code); ok {
			p.Code = r.FormValue("Code")
		}

		if err := n.Refresh(); err != nil {
			log.Printf("Unable to refresh node: %v", err)
			http.Error(w, "Unable to refresh node", http.StatusBadRequest)
			return
		}

		if nm == n.Name {
			break
		}

		if n.Name != "" {
			delete(g.Nodes, n.Name)
		}
		n.Name = nm
		g.Nodes[nm] = n

		q := url.Values{"node": []string{nm}}
		u := *r.URL
		u.RawQuery = q.Encode()
		log.Printf("redirecting to %v", u)
		http.Redirect(w, r, u.String(), http.StatusSeeOther) // should cause GET
		return
	}

	if err := renderNodeEditor(w, g, n); err != nil {
		log.Printf("Could not render source editor: %v", err)
		http.Error(w, "Could not render source editor", http.StatusInternalServerError)
		return
	}
	return
}
