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
	"io"

	"github.com/google/shenzhen-go/model"
)

// TODO: Replace these cobbled-together UIs with Polymer or something.
const nodeEditorTemplateSrc = `{{with .Node -}}
<head>
	<title>{{if .Name}}{{.Name}}{{else}}[New]{{end}}</title>
	<link type="text/css" rel="stylesheet" href="/.static/fonts.css">
	<link type="text/css" rel="stylesheet" href="/.static/main.css">
	<script>
	function showhidehelp(l) {
		d = document.getElementById('help');
		if (d.style.display == 'none') {
			d.style.display = 'block';
			l.innerText = 'Hide Help';
		} else {
			d.style.display = 'none';
			l.innerText = 'Show Help';
		}
	}
	</script>
</head>
<body>
	<h1>{{if .Name}}{{.Name}}{{else}}[New]{{end}}</h1>
	{{if .Name -}}
	<a href="?node={{.Name}}&clone" title="Make a copy of this goroutine.">Clone</a> | 
	{{if ne .Part.TypeKey "Code" -}}
	<a href="?node={{.Name}}&convert" class="destructive" title="Change this goroutine into a Code goroutine; it cannot be converted back.">Convert to Code</a> |
	{{end -}}
	<a href="?node={{.Name}}&delete" class="destructive" title="Delete this goroutine">Delete</a> | 
	{{end -}}
	Part type: {{.Part.TypeKey}} | 
	<a id="helplink" href="javascript:void(0);" onclick="showhidehelp(this);">Show Help</a>
	<div id="help" style="display:none">
		<h3>{{.Part.TypeKey}}</h3>
		{{.Part.Help}}
	</div>
	<form method="post">
		<div class="formfield">
			<label for="Name">Name</label>
			<input name="Name" type="text" required value="{{.Name}}">
		</div>
		<div class="formfield">
			<label for="Multiplicity">Multiplicity</label>
			<input name="Multiplicity" type="number" required pattern="^[1-9][0-9]*$" title="Must be a whole number, at least 1." value="{{if .Multiplicity}}{{.Multiplicity}}{{else}}1{{end}}">
		</div>
		<div class="formfield">
			<label for="Wait">Wait for this to finish</label>
			<input name="Wait" type="checkbox" {{if .Wait}}checked{{end}}>
		</div>
		{{template "part_view" $ }}
		<div class="formfield hcentre">
			<input type="submit" value="Save">
			<input type="button" value="Return" onclick="window.location.href='?'">
		</div>
	</form>
</body>
{{- end}}`

var nodeEditorTemplate = template.Must(template.New("nodeEditor").Parse(nodeEditorTemplateSrc))

// Node displays the node editor for a particular node.
func Node(w io.Writer, g *model.Graph, n *model.Node) error {
	t, err := nodeEditorTemplate.Clone()
	if err != nil {
		return err
	}
	if err := n.Part.AssociateEditor(t); err != nil {
		return err
	}
	return t.Execute(w, &struct {
		*model.Graph
		*model.Node
	}{g, n})
}
