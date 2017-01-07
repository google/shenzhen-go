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
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/google/shenzhen-go/graph"
)

const (
	graphEditorTemplateSrc = `<head>
	<title>{{$.Graph.Name}}</title><style>` + css + `</style>
</head>
<body>
<h1>{{$.Graph.Name}}</h1>
<div>
	<a href="?props">Properties</a> | 
	<a href="?save">Save</a> | 
	<a href="?build">Build</a> | 
	<a href="?run">Run</a> | 
	New: <a href="?node=new">Goroutine</a> <a href="?channel=new">Channel</a> | 
	View as: <a href="?go">Go</a> <a href="?dot">Dot</a> <a href="?json">JSON</a>
	<br><br>
	{{$.Diagram}}
</div>
</body>`

	// TODO: Replace these cobbled-together UIs with Polymer or something.
	graphPropertiesTemplateSrc = `<head>
	<title>{{.Name}}</title><style>` + css + `</style>
</head>
<body>
<h1>{{.Name}} Properties</h1>
{{.SourcePath}}
<div>
    <form method="post">
		<div class="formfield">
		    <label for="Name">Name</label>
			<input name="Name" type="text" required value="{{.Name}}">
		</div>
		<div class="formfield">
		    <label for="PackagePath">Package path</label>
			<input name="PackagePath" type="text" required value="{{.PackagePath}}">
		</div>
		<div class="formfield">
		    <label for="Imports">Imports</label>
			<textarea name="Imports" rows="10" cols="36">
				{{- range .Imports}}{{.}}{{"\n"}}{{end -}}
			</textarea>
		</div>
		<div class="formfield hcentre">
		    <input type="submit" value="Save">
			<input type="button" value="Return" onclick="window.location.href='?'">
		</div>
	</form>
</div>
</body>`
)

var (
	graphEditorTemplate     = template.Must(template.New("graphEditor").Parse(graphEditorTemplateSrc))
	graphPropertiesTemplate = template.Must(template.New("graphProperties").Parse(graphPropertiesTemplateSrc))
)

// Graph handles displaying/editing a graph.
func Graph(g *graph.Graph, w http.ResponseWriter, r *http.Request) {
	log.Printf("%s graph: %s", r.Method, r.URL)
	q := r.URL.Query()

	if _, t := q["props"]; t {
		if err := handlePropsRequest(g, w, r); err != nil {
			log.Printf("Could not execute graph properties editor template: %v", err)
			http.Error(w, "Could not execute graph properties editor template", http.StatusInternalServerError)
		}
		return
	}
	if _, t := q["dot"]; t {
		outputDotSrc(g, w)
		return
	}
	if _, t := q["go"]; t {
		outputGoSrc(g, w)
		return
	}
	if _, t := q["json"]; t {
		outputJSON(g, w)
		return
	}
	if _, t := q["build"]; t {
		if err := g.Build(); err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error building:\n%v", err)
			return
		}
		u := *r.URL
		u.RawQuery = ""
		http.Redirect(w, r, u.String(), http.StatusFound)
		return
	}
	if _, t := q["run"]; t {
		w.Header().Set("Content-Type", "text/plain")
		if err := g.Run(w, w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error building or running:\n%v", err)
		}
		return
	}
	if _, t := q["save"]; t {
		if err := g.SaveJSONFile(); err != nil {
			log.Printf("Failed to save JSON file: %v", err)
		}
		u := *r.URL
		u.RawQuery = ""
		http.Redirect(w, r, u.String(), http.StatusFound)
		return
	}
	if n := q["node"]; len(n) == 1 {
		Node(g, n[0], w, r)
		return
	}
	if n := q["channel"]; len(n) == 1 {
		Channel(g, n[0], w, r)
		return
	}

	var dot, svg bytes.Buffer
	if err := g.WriteDotTo(&dot); err != nil {
		log.Printf("Could not render to dot: %v", err)
		http.Error(w, "Could not render to dot", http.StatusInternalServerError)
		return
	}
	if err := dotToSVG(&svg, &dot); err != nil {
		log.Printf("Could not render dot to SVG: %v", err)
		http.Error(w, "Could not render dot to SVG", http.StatusInternalServerError)
		return
	}
	d := &struct {
		Diagram template.HTML
		Graph   *graph.Graph
	}{
		Diagram: template.HTML(svg.String()),
		Graph:   g,
	}
	if err := graphEditorTemplate.Execute(w, d); err != nil {
		log.Printf("Could not execute graph editor template: %v", err)
		http.Error(w, "Could not execute graph editor template", http.StatusInternalServerError)
	}
}

func handlePropsRequest(g *graph.Graph, w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "POST":
		return handlePropsPost(g, w, r)
	case "GET":
		return graphPropertiesTemplate.Execute(w, g)
	default:
		return fmt.Errorf("unsupported verb %q", r.Method)
	}
}

func handlePropsPost(g *graph.Graph, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	// Validate.
	nm := strings.TrimSpace(r.FormValue("Name"))
	if nm == "" {
		return fmt.Errorf(`name is empty [%q == ""]`, nm)
	}
	pp := strings.TrimSpace(r.FormValue("PackagePath"))
	if pp == "" {
		return fmt.Errorf(`package path is empty [%q == ""]`, pp)
	}

	imps := strings.Split(r.FormValue("Imports"), "\n")
	i := 0
	for _, imp := range imps {
		imp = strings.TrimSpace(imp)
		if imp == "" {
			continue
		}
		imps[i] = imp
		i++
	}
	imps = imps[:i]

	// Update.
	g.Name = nm
	g.PackagePath = pp
	g.Imports = imps

	return graphPropertiesTemplate.Execute(w, g)
}

func outputDotSrc(g *graph.Graph, w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Type", "text/vnd.graphviz")
	if err := g.WriteDotTo(w); err != nil {
		log.Printf("Could not render to dot: %v", err)
		http.Error(w, "Could not render to dot", http.StatusInternalServerError)
	}
}

func outputGoSrc(g *graph.Graph, w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Type", "text/golang")
	if err := g.WriteGoTo(w); err != nil {
		log.Printf("Could not render to Go: %v", err)
		http.Error(w, "Could not render to Go", http.StatusInternalServerError)
	}
}

func outputJSON(g *graph.Graph, w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Type", "application/json")
	if err := g.WriteJSONTo(w); err != nil {
		log.Printf("Could not encode JSON: %v", err)
		http.Error(w, "Could not encode JSON", http.StatusInternalServerError)
		return
	}
}
