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

package view

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var browseTemplate = template.Must(template.New("browse").Parse(string(templateResources["templates/browse.html"])))

// DirectoryEntry represents a file or directory in a filesystem.
type DirectoryEntry struct {
	IsDir bool
	Path  string
	Name  string
}

// Browse writes a filesystem browser to the response writer.
func Browse(w http.ResponseWriter, base string, dir []DirectoryEntry, params *Params) {
	d := &struct {
		Params  *Params
		Up      string
		Base    string
		Entries []DirectoryEntry
	}{
		Params:  params,
		Up:      filepath.Dir(base),
		Base:    base,
		Entries: dir,
	}
	if err := browseTemplate.Execute(w, d); err != nil {
		log.Printf("Could not execute browser template: %v", err)
		http.Error(w, "Could not execute browser template", http.StatusInternalServerError)
	}
}
