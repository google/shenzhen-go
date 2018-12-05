// Copyright 2017 Google Inc.
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

// Package server serves the user interface and API, and manages the data model.
package server

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/shenzhen-go/model"
	"github.com/google/shenzhen-go/server/view"
)

// New returns a new server.
func New(uiParams view.Params) *server {
	return &server{
		uiParams:     &uiParams,
		loadedGraphs: make(map[string]*serveGraph),
	}
}

type server struct {
	uiParams     *view.Params
	loadedGraphs map[string]*serveGraph
	sync.Mutex
}

func (c *server) lookupGraph(key string) (*serveGraph, error) {
	c.Lock()
	defer c.Unlock()
	g := c.loadedGraphs[key]
	if g == nil {
		return nil, status.Errorf(codes.NotFound, "graph %q not loaded", key)
	}
	return g, nil
}

func (c *server) createGraph(key string, graph *model.Graph) (*serveGraph, error) {
	c.Lock()
	defer c.Unlock()
	if c.loadedGraphs[key] != nil {
		return nil, status.Errorf(codes.NotFound, "graph %q already created", key)
	}
	sg := &serveGraph{Graph: graph}
	c.loadedGraphs[key] = sg
	return sg, nil
}

type serveGraph struct {
	*model.Graph
	sync.Mutex
}

func (sg *serveGraph) reload() error {
	f, err := os.Open(sg.Graph.FilePath)
	if err != nil {
		return status.Errorf(codes.NotFound, "open: %v", err)
	}
	defer f.Close()
	g, err := model.LoadJSON(f, sg.Graph.FilePath, sg.Graph.URLPath)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "load from JSON: %v", err)
	}
	sg.Graph = g
	return nil
}

func (sg *serveGraph) lookupChannel(channel string) (*model.Channel, error) {
	ch := sg.Channels[channel]
	if ch == nil {
		return nil, status.Errorf(codes.NotFound, "no such channel %q", channel)
	}
	return ch, nil
}

func (sg *serveGraph) lookupNode(node string) (*model.Node, error) {
	n := sg.Nodes[node]
	if n == nil {
		return nil, status.Errorf(codes.NotFound, "no such node %q", node)
	}
	return n, nil
}

func (c *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s browse: %s", r.Method, r.URL)

	if g, err := c.lookupGraph(r.URL.Path); err == nil {
		renderGraph(g, w, r, c.uiParams)
		return
	}

	filePath := filepath.FromSlash(r.URL.Path) // remap URL to filesystem separators
	base := filepath.Join(".", filePath)
	f, err := os.Open(base)
	if err != nil {
		log.Printf("Couldn't open: %v", err)
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		log.Printf("Couldn't stat: %v", err)
		http.NotFound(w, r)
		return
	}
	if !fi.IsDir() {
		g, err := model.LoadJSON(f, base, r.URL.Path)
		if err != nil {
			log.Printf("Not a directory or a valid JSON-encoded graph: %v", err)
			http.ServeContent(w, r, f.Name(), fi.ModTime(), f)
			return
		}
		sg, err := c.createGraph(r.URL.Path, g)
		if err != nil {
			log.Printf("Graph already created in server: %v", err)
			http.ServeContent(w, r, f.Name(), fi.ModTime(), f)
			return
		}
		renderGraph(sg, w, r, c.uiParams)
		return
	}

	if nu := r.URL.Query().Get("new"); nu != "" {
		// Check for an existing file.
		nfp := filepath.Join(base, nu)
		if _, err := os.Stat(nfp); !os.IsNotExist(err) {
			log.Printf("Asked to create %q but it already exists", nfp)
			http.Error(w, "File already exists", http.StatusNotModified)
			return
		}
		pkgp, err := GuessPackagePath(nfp)
		if err != nil {
			log.Printf("Guessing a package path: %v", err)
		}
		urlPath := path.Join(r.URL.Path, nu)
		if _, err := c.createGraph(urlPath, model.NewGraph(nfp, urlPath, pkgp)); err != nil {
			log.Printf("Graph already created in server: %v", err)
		} else {
			log.Printf("Created new graph: %v", nfp)
		}
		http.Redirect(w, r, urlPath, http.StatusSeeOther)
		return
	}

	fis, err := f.Readdir(0)
	if err != nil {
		log.Printf("Couldn't readdir: %s", err)
		http.NotFound(w, r)
		return
	}

	var e []view.DirectoryEntry
	for _, fi := range fis {
		if strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		e = append(e, view.DirectoryEntry{
			IsDir: fi.IsDir(),
			Name:  fi.Name(),
			Path:  path.Join(r.URL.Path, fi.Name()),
		})
	}

	sort.Slice(e, func(i, j int) bool {
		return e[i].Name < e[j].Name
	})

	view.Browse(w, base, e, c.uiParams)
}
