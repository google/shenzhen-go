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

package server

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/server/view"
)

func (c *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s browse: %s", r.Method, r.URL)

	if g, err := c.lookupGraph(r.URL.Path); err == nil {
		renderGraph(g, w, r)
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
		renderGraph(sg, w, r)
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

	view.Browse(w, base, e)
}
