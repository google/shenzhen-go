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
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"shenzhen-go/graph"
	"shenzhen-go/parts"
)

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

	switch r.Method {
	case "POST":
		if n == nil {
			n = &graph.Node{
				Part: &parts.Code{},
			}
		}

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

		delete(g.Nodes, n.Name)
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
