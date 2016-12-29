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
	"regexp"
	"strconv"

	"shenzhen-go/graph"
)

const channelEditorTemplateSrc = `<head>
	<title>{{with .Channel}}{{.Name}}{{else}}[New]{{end}}</title><style>` + css + `</style>
</head>
<body>
	<h1>{{with .Channel}}{{.Name}}{{else}}[New]{{end}}</h1>
	<form method="post">
		<div class="formfield"><label for="Name">Name</label><input type="text" name="Name" required pattern="^[_a-zA-Z][_a-zA-Z0-9]*$" title="Must start with a letter or underscore, and only contain letters, digits, or underscores." value="{{with .Channel}}{{.Name}}{{end}}"></div>
		<div class="formfield"><label for="Type">Type</label><input type="text" name="Type" required value="{{with .Channel}}{{.Type}}{{end}}"></div>
		<div class="formfield"><label for="Cap">Capacity</label><input type="text" name="Cap" required pattern="^[0-9]+$" title="Must be a whole number, at least 0." value="{{with .Channel}}{{.Cap}}{{end}}"></div>
		<div class="formfield hcentre"><input type="submit" value="Save"> <input type="button" value="Return" onclick="window.location.href='?'"></div>
	</form>
</body>`

var (
	channelEditorTemplate = template.Must(template.New("channelEditor").Parse(channelEditorTemplateSrc))

	identifierRE = regexp.MustCompile(`^[_a-zA-Z][_a-zA-Z0-9]*$`)
)

func renderChannelEditor(dst io.Writer, g *graph.Graph, e *graph.Channel) error {
	return channelEditorTemplate.Execute(dst, struct {
		Graph   *graph.Graph
		Channel *graph.Channel
	}{g, e})
}

// Channel handles viewing/editing a channel.
func Channel(g *graph.Graph, name string, w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	e, found := g.Channels[name]
	if name != "new" && !found {
		http.Error(w, fmt.Sprintf("Channel %q not found", name), http.StatusNotFound)
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

		q := url.Values{
			"channel": []string{nn},
		}
		u := *r.URL
		u.RawQuery = q.Encode()
		log.Printf("redirecting to %v", u)
		http.Redirect(w, r, u.String(), http.StatusSeeOther) // should cause GET
		return
	}

	if err := renderChannelEditor(w, g, e); err != nil {
		log.Printf("Could not render source editor: %v", err)
		http.Error(w, "Could not render source editor", http.StatusInternalServerError)
		return
	}
	return
}
