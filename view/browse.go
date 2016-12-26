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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"shenzhen-go/graph"
)

// dirBrowser serves a way of visually navigating the filesystem.
type dirBrowser struct {
	loadedGraphs map[string]*graph.Graph
}

func NewBrowser() http.Handler {
	return &dirBrowser{
		loadedGraphs: make(map[string]*graph.Graph),
	}
}

type entry struct {
	IsDir bool
	Path  string
	Name  string
}

func (b *dirBrowser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s browse: %s", r.Method, r.URL)

	path := r.URL.Path
	if g, ok := b.loadedGraphs[path]; ok {
		Graph(g, w, r)
		return
	}

	base := filepath.Join(".", path)
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
		g, err := graph.LoadJSON(f, base)
		if err != nil {
			log.Printf("Not a directory or a valid JSON graph: %v", err)
			http.NotFound(w, r)
		}
		b.loadedGraphs[path] = g
		Graph(g, w, r)
		return
	}
	fis, err := f.Readdir(0)
	if err != nil {
		log.Printf("Couldn't readdir: %s", err)
		http.NotFound(w, r)
		return
	}

	var e []entry
	for _, fi := range fis {
		if strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		e = append(e, entry{
			IsDir: fi.IsDir(),
			Name:  fi.Name(),
			Path:  filepath.Join(path, fi.Name()),
		})
	}

	d := &struct {
		Base    string
		Entries []entry
	}{
		Base:    base,
		Entries: e,
	}
	if err := browseTemplate.Execute(w, d); err != nil {
		log.Printf("Could not execute browser template: %v", err)
		http.Error(w, "Could not execute browser template", http.StatusInternalServerError)
	}
}
