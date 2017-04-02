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

package controller

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/model/parts"
	"github.com/google/shenzhen-go/view"
)

// node handles viewing/editing a node.
func handleNodeRequest(g *model.Graph, name string, w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	q := r.URL.Query()
	_, clone := q["clone"]
	_, convert := q["convert"]
	_, del := q["delete"]

	var n *model.Node
	if name == "new" {
		if clone || convert || del {
			http.Error(w, "Asked for a new node, but also to clone/convert/delete the node", http.StatusBadRequest)
			return
		}
		t := q.Get("type")
		pf, ok := model.PartFactories[t]
		if !ok {
			http.Error(w, "Asked for a new node, but didn't supply a valid type", http.StatusBadRequest)
			return
		}
		n = &model.Node{
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
	case convert:
		h, b, t := n.Part.Impl()
		n.Part = &parts.Code{
			Head: h,
			Body: b,
			Tail: t,
		}
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
		err = view.Node(w, g, n)
	default:
		err = fmt.Errorf("unsupported verb %q", r.Method)
	}

	if err != nil {
		msg := fmt.Sprintf("Could not handle request: %v", err)
		log.Printf(msg)
		http.Error(w, msg, http.StatusInternalServerError)
	}
}

func handleNodePost(g *model.Graph, n *model.Node, w http.ResponseWriter, r *http.Request) error {
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
	part := model.PartFactories[n.Part.TypeKey()]()
	if err := part.Update(r); err != nil {
		return err
	}

	// Update.
	n.Multiplicity = uint(mult)
	n.Wait = (r.FormValue("Wait") == "on")
	n.Part = part

	// No name change? No need to readjust the map or redirect.
	// So render the usual editor.
	if nm == n.Name {
		return view.Node(w, g, n)
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
