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
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/google/shenzhen-go/graph"
)

// TODO: Replace these cobbled-together UIs with Polymer or something.
const channelEditorTemplateSrc = `<head>
	<title>{{if .Name}}{{.Name}}{{else}}[New]{{end}}</title>
	<link type="text/css" rel="stylesheet" href="/.static/fonts.css">
	<link type="text/css" rel="stylesheet" href="/.static/main.css">
</head>
<body>
	<h1>{{if .Name}}{{.Name}}{{else}}[New]{{end}}</h1>
	{{if .Name}}
	<a href="?channel={{.Name}}&clone">Clone</a> | 
	<a href="?channel={{.Name}}&delete">Delete</a>
	{{end}}
	<form method="post">
		<div class="formfield">
			<label for="Name">Name</label>
			<input type="text" name="Name" required pattern="^[_a-zA-Z][_a-zA-Z0-9]*$" title="Must start with a letter or underscore, and only contain letters, digits, or underscores." value="{{.Name}}">
		</div>
		<div class="formfield">
			<label for="Type">Type</label>
			<input type="text" name="Type" required value="{{.Type}}">
		</div>
		<div class="formfield">
			<label for="Cap">Capacity</label>
			<input type="number" name="Cap" required pattern="^[0-9]+$" title="Must be a whole number, at least 0." value="{{.Cap}}">
		</div>
		<div class="formfield hcentre">
			<input type="submit" value="Save">
			<input type="button" value="Return" onclick="window.location.href='?'">
		</div>
	</form>
</body>`

var (
	channelEditorTemplate = template.Must(template.New("channelEditor").Parse(channelEditorTemplateSrc))

	identifierRE = regexp.MustCompile(`^[_a-zA-Z][_a-zA-Z0-9]*$`)
)

// Channel handles viewing/editing a channel.
func Channel(g *graph.Graph, name string, w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	q := r.URL.Query()
	_, clone := q["clone"]
	_, del := q["delete"]

	var e *graph.Channel
	if name == "new" {
		if clone || del {
			http.Error(w, "Asked for a new channel, but also to clone or delete the channel", http.StatusBadRequest)
			return
		}
		e = new(graph.Channel)
	} else {
		e1, ok := g.Channels[name]
		if !ok {
			http.Error(w, fmt.Sprintf("Channel %q not found", name), http.StatusNotFound)
			return
		}
		e = e1
	}

	switch {
	case clone:
		e2 := *e
		e2.Name = ""
		e = &e2
	case del:
		delete(g.Channels, e.Name)
		u := *r.URL
		u.RawQuery = ""
		log.Printf("redirecting to %v", &u)
		http.Redirect(w, r, u.String(), http.StatusSeeOther) // should cause GET
		return
	}

	var err error
	switch r.Method {
	case "POST":
		err = handleChannelPost(g, e, w, r)
	case "GET":
		err = channelEditorTemplate.Execute(w, e)
	default:
		err = fmt.Errorf("unsupported verb %q", r.Method)
	}

	if err != nil {
		log.Printf("Could not handle request: %v", err)
		http.Error(w, "Could not handle request", http.StatusInternalServerError)
	}
}

func handleChannelPost(g *graph.Graph, e *graph.Channel, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	// Validate.
	nn := r.FormValue("Name")
	if !identifierRE.MatchString(nn) {
		return fmt.Errorf("invalid name [%q !~ %q]", nn, identifierRE)
	}

	ci, err := strconv.Atoi(r.FormValue("Cap"))
	if err != nil {
		return err
	}
	if ci < 0 {
		return fmt.Errorf("invalid capacity [%d < 0]", ci)
	}

	// Update.
	e.Type = r.FormValue("Type")
	e.Cap = ci

	// No name change? No need to readjust the map or redirect.
	// So render the usual editor.
	if nn == e.Name {
		return channelEditorTemplate.Execute(w, e)
	}

	// Do name changes last since they cause a redirect.
	if e.Name != "" {
		for _, n := range g.Nodes {
			cr, cw := n.Channels()
			if cr.Ni(e.Name) || cw.Ni(e.Name) {
				n.RenameChannel(e.Name, nn)
			}
		}
		delete(g.Channels, e.Name)
	}
	e.Name = nn
	g.Channels[nn] = e

	q := url.Values{
		"channel": []string{nn},
	}
	u := *r.URL
	u.RawQuery = q.Encode()
	log.Printf("redirecting to %v", u)
	http.Redirect(w, r, u.String(), http.StatusSeeOther) // should cause GET
	return nil
}
