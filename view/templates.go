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
)

const (
	css = `
	body {
		font-family: "Go","San Francisco","Helvetica Neue",Helvetica,sans-serif;
		float: none;
		max-width: 800px;
		margin: 20 auto 0;
	}
	form {
		float: none;
		max-width: 800px;
		margin: 0 auto;
	}
	div.formfield {
		margin-top: 12px;
		margin-bottom: 12px;
	}
	label {
		float: left;
		text-align: right;
		margin-right: 15px;
		width: 50%;
	}
	input {
		font-family: "Go Mono","Fira Code",sans-serif;
		font-size: 12pt;
	}
	select {
		font-family: "Go Mono","Fira Code",sans-serif;
		font-size: 12pt;
	}
	textarea {
		font-family: "Go Mono","Fira Code",sans-serif;
		font-size: 12pt;
	}
	div svg {
		display: block;
		margin: 0 auto;
	}
	div.hcentre {
		text-align: center;
	}
	table.browse {
		font-family: "Go Mono","Fira Code",sans-serif;
		font-size: 12pt;
	}
	`

	channelEditorTemplateSrc = `<head>
	<title>{{with .Channel}}{{.Name}}{{else}}[New]{{end}}</title><style>` + css + `</style>
</head>
<body>
	<h1>{{with .Channel}}{{.Name}}{{else}}[New]{{end}}</h1>
	<form method="post">
		<div class="formfield"><label for="Name">Name</label><input type="text" name="Name" required pattern="^[_a-zA-Z][_a-zA-Z0-9]*$" title="Must start with a letter or underscore, and only contain letters, digits, or underscores." value="{{with .Channel}}{{.Name}}{{end}}"></div>
		<div class="formfield"><label for="Type">Type</label><input type="text" name="Type" required value="{{with .Channel}}{{.Type}}{{end}}"></div>
		<div class="formfield"><label for="Cap">Capacity</label><input type="text" name="Cap" required value="{{with .Channel}}{{.Cap}}{{end}}"></div>
		<div class="formfield hcentre"><input type="submit" value="Save"> <input type="button" value="Return" onclick="window.location.href='?'"></div>
	</form>
</body>`

	nodeEditorTemplateSrc = `<head>
	<title>{{.Node.Name}}</title><style>` + css + `</style>
</head>
<body>
	{{with .Node}}
	<h1>{{.Name}}</h1>
	<form method="post">
		<div class="formfield"><label for="Name">Name</label><input name="Name" type="text" required value="{{.Name}}"></div>
		<div class="formfield"><label for="Wait">Wait for this to finish</label><input name="Wait" type="checkbox" {{if .Wait}}checked{{end}}></div>
		<div class="formfield"><textarea name="Code" rows="25" cols="80">{{.Impl}}</textarea></div>
		<div class="formfield hcentre"><input type="submit" value="Save"> <input type="button" value="Return" onclick="window.location.href='?'"></div>
	</form>
	{{end}}
</body>`

	browseTemplateSrc = `<head>
	<title>SHENZHEN GO</title><style>` + css + `</style>
</head>
<body>
<h1>SHENZHEN GO</h1>
<div>
<h2>{{$.Base}}</h2>
<table class="browse">
{{range $.Entries}}
<tr><td>{{if .IsDir}}&lt;dir&gt;{{end}}</td><td><a href="/{{.Path}}">{{.Name}}</a></td></tr>{{end}}
</table>
</div>
</body>`

	graphEditorTemplateSrc = `<head>
	<title>{{$.Graph.Name}}</title><style>` + css + `</style>
</head>
<body>
<h1>{{$.Graph.Name}}</h1>
<div>View as: <a href="?go">Go</a> <a href="?dot">Dot</a> <a href="?json">JSON</a> | <a href="?run">Run</a><br>
{{$.Diagram}}
</div>
</body>`
)

var (
	browseTemplate        = template.Must(template.New("browse").Parse(browseTemplateSrc))
	graphEditorTemplate   = template.Must(template.New("graphEditor").Parse(graphEditorTemplateSrc))
	nodeEditorTemplate    = template.Must(template.New("nodeEditor").Parse(nodeEditorTemplateSrc))
	channelEditorTemplate = template.Must(template.New("channelEditor").Parse(channelEditorTemplateSrc))
)
