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
)

// TODO: Replace these cobbled-together UIs with Polymer or something.
// TODO: Some way of deleting nodes.
const nodeEditorTemplateSrc = `{{with .Node -}}
<head>
	<title>{{if .Name}}{{.Name}}{{else}}[New]{{end}}</title>
	<link type="text/css" rel="stylesheet" href="/.static/fonts.css">
	<link type="text/css" rel="stylesheet" href="/.static/main.css">
	<script>
	function showhidehelp(l) {
		d = document.getElementById('help');
		if (d.style.display == 'none') {
			d.style.display = 'block';
			l.innerText = 'Hide Help';
		} else {
			d.style.display = 'none';
			l.innerText = 'Show Help';
		}
	}
	</script>
</head>
<body>
	<h1>{{if .Name}}{{.Name}}{{else}}[New]{{end}}</h1>
	{{if .Name}}
	<a href="?node={{.Name}}&clone">Clone</a> | 
	<a href="?node={{.Name}}&delete">Delete</a> | 
	{{end}}
	Part type: {{.Part.TypeKey}} | 
	<a id="helplink" href="javascript:void(0);" onclick="showhidehelp(this);">Show Help</a>
	<div id="help" style="display:none">
		<h3>{{.Part.TypeKey}}</h3>
		{{.Part.Help}}
	</div>
	<form method="post">
		<div class="formfield">
			<label for="Name">Name</label>
			<input name="Name" type="text" required value="{{.Name}}">
		</div>
		<div class="formfield">
			<label for="Multiplicity">Multiplicity</label>
			<input name="Multiplicity" type="number" required pattern="^[1-9][0-9]*$" title="Must be a whole number, at least 1." value="{{if .Multiplicity}}{{.Multiplicity}}{{else}}1{{end}}">
		</div>
		<div class="formfield">
			<label for="Wait">Wait for this to finish</label>
			<input name="Wait" type="checkbox" {{if .Wait}}checked{{end}}>
		</div>
		{{template "part_view" $ }}
		<div class="formfield hcentre">
			<input type="submit" value="Save">
			<input type="button" value="Return" onclick="window.location.href='?'">
		</div>
	</form>
</body>
{{- end}}`

var nodeEditorTemplate = template.Must(template.New("nodeEditor").Parse(nodeEditorTemplateSrc))

// Node handles viewing/editing a node.
func Node(g *graph.Graph, name string, w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	q := r.URL.Query()
	_, clone := q["clone"]
	_, del := q["delete"]

	var n *graph.Node
	if name == "new" {
		if clone || del {
			http.Error(w, "Asked for a new node, but also to clone or delete the node", http.StatusBadRequest)
			return
		}
		t := q.Get("type")
		pf, ok := graph.PartFactories[t]
		if !ok {
			http.Error(w, "Asked for a new node, but didn't supply a valid type", http.StatusBadRequest)
			return
		}
		n = &graph.Node{
			Part: pf(),
			Wait: true,
		}
	} else {
		n1, found := g.Nodes[name]
		if !found {
			http.Error(w, fmt.Sprintf("Node %q not found", name), http.StatusNotFound)
			return
		}
		n = n1
	}

	switch {
	case clone:
		n = n.Copy()
	case del:
		delete(g.Nodes, n.Name)
		u := *r.URL
		u.RawQuery = ""
		log.Printf("redirecting to %v", &u)
		http.Redirect(w, r, u.String(), http.StatusSeeOther) // should cause GET
		return
	}

	var err error
	switch r.Method {
	case "POST":
		err = handleNodePost(g, n, w, r)
	case "GET":
		err = renderNodeEditor(w, g, n)
	default:
		err = fmt.Errorf("unsupported verb %q", r.Method)
	}

	if err != nil {
		msg := fmt.Sprintf("Could not handle request: %v", err)
		log.Printf(msg)
		http.Error(w, msg, http.StatusInternalServerError)
	}
}

func renderNodeEditor(dst io.Writer, g *graph.Graph, n *graph.Node) error {
	t, err := nodeEditorTemplate.Clone()
	if err != nil {
		return err
	}
	if err := n.Part.AssociateEditor(t); err != nil {
		return err
	}
	return t.Execute(dst, &struct {
		*graph.Graph
		*graph.Node
	}{g, n})
}

func handleNodePost(g *graph.Graph, n *graph.Node, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	// Validate.
	nm := strings.TrimSpace(r.FormValue("Name"))
	if nm == "" {
		return fmt.Errorf(`name is empty [%q == ""]`, nm)
	}

	mult, err := strconv.Atoi(r.FormValue("Multiplicity"))
	if err != nil {
		return err
	}
	if mult < 1 {
		return fmt.Errorf("multiplicity too small [%d < 1]", mult)
	}

	// Create a new part of the same type, and update it.
	// This ensures the settings for the part are valid before
	// updating the node.
	part := graph.PartFactories[n.Part.TypeKey()]()
	if err := part.Update(r); err != nil {
		return err
	}

	// Update.
	n.Multiplicity = uint(mult)
	n.Wait = (r.FormValue("Wait") == "on")
	n.Part = part
	if err := g.RecomputeNode(n); err != nil {
		log.Printf("Recomputing node %s: %v", nm, err)
	}

	// No name change? No need to readjust the map or redirect.
	// So render the usual editor.
	if nm == n.Name {
		return renderNodeEditor(w, g, n)
	}

	// Do name changes last since they cause a redirect.
	if n.Name != "" {
		delete(g.Nodes, n.Name)
	}
	n.Name = nm
	g.Nodes[nm] = n

	q := url.Values{"node": []string{nm}}
	u := *r.URL
	u.RawQuery = q.Encode()
	log.Printf("redirecting to %v", &u)
	http.Redirect(w, r, u.String(), http.StatusSeeOther) // should cause GET
	return nil
}
