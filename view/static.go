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
	"log"
	"net/http"

	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/font/gofont/gomediumitalic"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/gofont/gomonobolditalic"
	"golang.org/x/image/font/gofont/gomonoitalic"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	fontsCSS = `@font-face {
	font-family: 'Go';
	src: url('/.static/fonts/GoMedium-Italic.ttf') format('truetype');
	font-weight: 500;
	font-style: italic;
}

@font-face {
	font-family: 'Go';
	src: url('/.static/fonts/Go-Italic.ttf') format('truetype');
	font-weight: normal;
	font-style: italic;
}

@font-face {
	font-family: 'Go';
	src: url('/.static/fonts/Go-Bold.ttf') format('truetype');
	font-weight: bold;
	font-style: normal;
}

@font-face {
	font-family: 'Go';
	src: url('/.static/fonts/GoMedium.ttf') format('truetype');
	font-weight: 500;
	font-style: normal;
}

@font-face {
	font-family: 'Go';
	src: url('/.static/fonts/Go-BoldItalic.ttf') format('truetype');
	font-weight: bold;
	font-style: italic;
}

@font-face {
	font-family: 'Go';
	src: url('/.static/fonts/GoRegular.ttf') format('truetype');
	font-weight: normal;
	font-style: normal;
}

@font-face {
	font-family: 'Go Mono';
	src: url('/.static/fonts/GoMono-Bold.ttf') format('truetype');
	font-weight: bold;
	font-style: normal;
}

@font-face {
	font-family: 'Go Mono';
	src: url('/.static/fonts/GoMono.ttf') format('truetype');
	font-weight: normal;
	font-style: normal;
}

@font-face {
	font-family: 'Go Mono';
	src: url('/.static/fonts/GoMono-Italic.ttf') format('truetype');
	font-weight: normal;
	font-style: italic;
}

@font-face {
	font-family: 'Go Mono';
	src: url('/.static/fonts/GoMono-BoldItalic.ttf') format('truetype');
	font-weight: bold;
	font-style: italic;
}`

	mainCSS = `body {
	font-family: Go,'San Francisco','Helvetica Neue',Helvetica,Arial,sans-serif;
	float: none;
	max-width: 800px;
	margin: 20 auto 0;
}
a:link, a:visited {
	color: #05d;
	text-decoration: none;
}
a:hover {
	color: #07f;
	text-decoration: underline;
}
a.destructive:link, a.destructive:visited {
    color: #d03;
}
a.destructive:hover {
    color: #f04;
}
code {
	font-family: 'Go Mono','Fira Code',Menlo,sans-serif;
	color: #066;
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
	width: 30%;
}
input {
	font-family: 'Go Mono','Fira Code',Menlo,sans-serif;
	font-size: 12pt;
}
input[type=text] {
	width: 65%;
}
select {
	font-family: 'Go Mono','Fira Code',Menlo,sans-serif;
	font-size: 12pt;
}
textarea {
	font-family: 'Go Mono','Fira Code',Menlo,sans-serif;
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
	font-family: 'Go Mono','Fira Code',Menlo,sans-serif;
	font-size: 12pt;
	margin-top: 16pt;
}
fieldset {
	margin: 4px;
}
fieldset#pathtemplate {
	display: none;
}
.dropdown {
    position: relative;
    display: inline-block;
}
.dropdown-content {
    display: none;
    position: absolute;
    background-color: #fff;
    box-shadow: 0px 6px 12px 0px rgba(0,0,0,0.2);
    padding: 4px 4px;
    z-index: 1;
}
.dropdown:hover .dropdown-content {
    display: block;
}
.dropdown-content ul {
	list-style-type: none;
   	margin: 0;
   	padding: 0;
   	overflow: hidden;
}
pre.codeedit {
	font-family: 'Go Mono','Fira Code',Menlo,sans-serif;
	height: 600px;
    width: 100%;
}
`
)

var (
	staticMap = map[string][]byte{
		"fonts.css": []byte(fontsCSS),
		"main.css":  []byte(mainCSS),

		"fonts/GoMedium-Italic.ttf":   gomediumitalic.TTF,
		"fonts/Go-Italic.ttf":         goitalic.TTF,
		"fonts/Go-Bold.ttf":           gobold.TTF,
		"fonts/GoMedium.ttf":          gomedium.TTF,
		"fonts/Go-BoldItalic.ttf":     gobolditalic.TTF,
		"fonts/GoRegular.ttf":         goregular.TTF,
		"fonts/GoMono-Bold.ttf":       gomonobold.TTF,
		"fonts/GoMono.ttf":            gomono.TTF,
		"fonts/GoMono-Italic.ttf":     gomonoitalic.TTF,
		"fonts/GoMono-BoldItalic.ttf": gomonobolditalic.TTF,
	}
)

type staticHandler struct{}

// Static serves "static resources".
var Static staticHandler

func (staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s static: %s", r.Method, r.URL)
	d := staticMap[r.URL.Path]
	if d == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Add("Cache-Control", "public")
	w.Header().Add("Cache-Control", "max-age=86400")
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}
