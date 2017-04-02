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
	"html/template"
	"net/http"
	"regexp"

	"github.com/google/shenzhen-go/model"
)

// TODO: Replace these cobbled-together UIs with Polymer or something.
const channelEditorTemplateSrc = `<head>
	<title>Channel</title>
	<link type="text/css" rel="stylesheet" href="/.static/fonts.css">
	<link type="text/css" rel="stylesheet" href="/.static/main.css">
</head>
<body>
	<h1>Channel</h1>
	{{if not .New}}
	<a href="?channel={{.Index}}&clone">Clone</a> | 
	<a href="?channel={{.Index}}&delete">Delete</a>
	{{end}}
	<form method="post">
		<input type="hidden" name="New" value="{{.New}}>
		<div class="formfield">
			<label for="Type">Type</label>
			<input type="text" name="Type" required value="{{.Type}}">
		</div>
		<div class="formfield">
			<label for="Cap">Capacity</label>
			<input type="number" name="Cap" required pattern="^[0-9]+$" title="Must be a whole number, at least 0." value="{{.Cap}}">
		</div>
		<table>
			<thead>
				<tr>
					<th>Connection</th>
					<th></th>
				</tr>
			</thead>
			{{range $conn := .Connections -}}
			<tr>
				<td>
					<select>
					{{range $node := $.Graph.Nodes }}
					{{range $arg, $type := $node.InputArgs}}
					{{if eq $type $.Channel.Type}}
					{{$val := printf "%s.%s" $node $arg}}
						<option value="{{$val}}" {{if eq $val $conn.String}}selected{{end}}>{{$val}}</option>
					{{end -}}
					{{end -}}
					{{range $arg, $type := $node.OutputArgs}}
					{{if eq $type $.Channel.Type}}
					{{$val := printf "%s.%s" $node $arg}}
						<option value="{{$val}}" {{if eq $val $conn.String}}selected{{end}}>{{$val}}</option>
					{{end -}}
					{{end -}}
					{{end -}}
					</select>
				</td>
				<td><a href="javascript:void(0)" onclick="removeconn('{{$conn}}}')">Remove connection</td></tr>
			{{end}}
		</table>
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

// Channel displays the channel editor for a particular channel.
func Channel(w http.ResponseWriter, g *model.Graph, e *model.Channel, new bool) error {
	return channelEditorTemplate.Execute(w, &struct {
		*model.Graph
		*model.Channel
		New bool
	}{g, e, new})
}
