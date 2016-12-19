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

// Package view provides the user interface.
package view

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"shenzhen-go/graph"
	"shenzhen-go/parts"
)

var identifierRE = regexp.MustCompile(`^[_a-zA-Z][_a-zA-Z0-9]*$`)

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
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	if err := enc.Encode(g); err != nil {
		log.Printf("Could not encode JSON: %v", err)
		http.Error(w, "Could not encode JSON", http.StatusInternalServerError)
		return
	}
}

func renderNodeEditor(dst io.Writer, g *graph.Graph, n *graph.Node) error {
	return nodeEditorTemplate.Execute(dst, struct {
		Graph *graph.Graph
		Node  *graph.Node
	}{g, n})
}

func renderChannelEditor(dst io.Writer, g *graph.Graph, e *graph.Channel) error {
	return channelEditorTemplate.Execute(dst, struct {
		Graph   *graph.Graph
		Channel *graph.Channel
	}{g, e})
}

// HandleRootRequest handles the overview / root page requests.
func HandleRootRequest(g *graph.Graph, w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	q := r.URL.Query()
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
	if _, t := q["run"]; t {
		if err := g.SaveBuildAndRun(); err != nil {
			log.Printf("Failed to save, build, run: %v", err)
		}
	}

	var dot, svg bytes.Buffer
	if err := g.WriteDotTo(&dot); err != nil {
		log.Printf("Could not render to dot: %v", err)
		http.Error(w, "Could not render to dot", http.StatusInternalServerError)
	}
	if err := dotToSVG(&svg, &dot); err != nil {
		log.Printf("Could not render dot to SVG: %v", err)
		http.Error(w, "Could not render dot to SVG", http.StatusInternalServerError)
	}
	if err := rootTemplate.Execute(w, template.HTML(svg.String())); err != nil {
		log.Printf("Could not execute root template: %v", err)
		http.Error(w, "Could not execute root template", http.StatusInternalServerError)
	}
}

// HandleChannelRequest handles channel viewer/editor requests.
func HandleChannelRequest(g *graph.Graph, w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	nm := strings.TrimPrefix(r.URL.Path, "/channel/")

	e, found := g.Channels[nm]
	if nm != "new" && !found {
		http.Error(w, fmt.Sprintf("Channel %q not found", nm), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "POST":
		if e == nil {
			e = new(graph.Channel)
		}

		// Parse...
		if err := r.ParseForm(); err != nil {
			log.Printf("Could not parse form: %v", err)
			http.Error(w, "Could not parse", http.StatusBadRequest)
			return
		}

		// ...Validate...
		nn := r.FormValue("Name")
		if !identifierRE.MatchString(nn) {
			msg := fmt.Sprintf("Invalid identifier %q !~ %q", nn, identifierRE)
			log.Printf(msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		ci, err := strconv.Atoi(r.FormValue("Cap"))
		if err != nil {
			log.Printf("Capacity is not an integer: %v", err)
			http.Error(w, "Capacity is not an integer", http.StatusBadRequest)
			return
		}
		if ci < 0 {
			log.Printf("Must specify nonnegative capacity [%d < 0]", ci)
			http.Error(w, "Capacity must be non-negative", http.StatusBadRequest)
			return
		}

		// ...update...
		e.Type = r.FormValue("Type")
		e.Cap = ci

		// Do name changes last since they cause a redirect.
		if nn == e.Name {
			break
		}
		delete(g.Channels, e.Name)
		e.Name = nn
		g.Channels[nn] = e
		http.Redirect(w, r, "/channel/"+nn, http.StatusSeeOther) // should cause GET
		return
	}

	if err := renderChannelEditor(w, g, e); err != nil {
		log.Printf("Could not render source editor: %v", err)
		http.Error(w, "Could not render source editor", http.StatusInternalServerError)
		return
	}
	return
}

// HandleNodeRequest handles node viewer/editor requests.
func HandleNodeRequest(g *graph.Graph, w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	nm := strings.TrimPrefix(r.URL.Path, "/node/")
	n, found := g.Nodes[nm]
	if !found {
		http.Error(w, fmt.Sprintf("Node %q not found", nm), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "POST":
		// Update the node.
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

		if err := n.Refresh(g); err != nil {
			log.Printf("Unable to extract channels used from code: %v", err)
			http.Error(w, "Unable to extract channels used from code", http.StatusBadRequest)
			return
		}

		if nm == n.Name {
			break
		}

		delete(g.Nodes, n.Name)
		n.Name = nm
		g.Nodes[nm] = n

		http.Redirect(w, r, "/node/"+nm, http.StatusSeeOther) // should cause GET
		return
	}

	if err := renderNodeEditor(w, g, n); err != nil {
		log.Printf("Could not render source editor: %v", err)
		http.Error(w, "Could not render source editor", http.StatusInternalServerError)
		return
	}
	return
}

func pipeThru(dst io.Writer, cmd *exec.Cmd, src io.Reader) error {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := io.Copy(stdin, src); err != nil {
		return err
	}
	if err := stdin.Close(); err != nil {
		return err
	}
	if _, err := io.Copy(dst, stdout); err != nil {
		return err
	}
	return cmd.Wait()
}

func dotToSVG(dst io.Writer, src io.Reader) error {
	return pipeThru(dst, exec.Command(`dot`, `-Tsvg`), src)
}
