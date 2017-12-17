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

const browseTemplateSrc = `<head>
	<title>SHENZHEN GO</title>
	<link type="text/css" rel="stylesheet" href="/.static/fonts.css">
	<link type="text/css" rel="stylesheet" href="/.static/main.css">
</head>
<body>
	<div class="browse-container">
		<h1>SHENZHEN GO</h1>
		<h2>{{$.Base}}</h2>
		<a href="/{{.Up}}">Up</a> |
		<div class="dropdown"> 
			<a href="javascript:void(0)">New</a>
			<form method="GET" class="dropdown-content">
				<input type="text" name="new" required style="width:200px">
				<a href="javascript:void(0)" onclick="this.parentElement.submit();">Create</a>
			</form>
		</div>
		<table class="browse">
			{{range $.Entries -}}
			<tr>
				<td>{{if .IsDir}}&lt;dir&gt;{{end}}</td>
				<td><a href="{{.Path}}">{{.Name}}</a></td>
			</tr>
			{{- end}}
		</table>
	</div>
</body>`

var browseTemplate = template.Must(template.New("browse").Parse(browseTemplateSrc))

// DirectoryEntry represents a file or directory in a filesystem.
type DirectoryEntry struct {
	IsDir bool
	Path  string
	Name  string
}

// Browse writes a filesystem browser to the response writer.
func Browse(w http.ResponseWriter, base string, dir []DirectoryEntry) {
	d := &struct {
		Up      string
		Base    string
		Entries []DirectoryEntry
	}{
		Up:      filepath.Dir(base),
		Base:    base,
		Entries: dir,
	}
	if err := browseTemplate.Execute(w, d); err != nil {
		log.Printf("Could not execute browser template: %v", err)
		http.Error(w, "Could not execute browser template", http.StatusInternalServerError)
	}
}
