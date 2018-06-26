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
	"bytes"
	"log"
	"net/http"
	"time"

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

var staticMap = map[string][]byte{
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

func load(m map[string][]byte) {
	for k, v := range m {
		staticMap[k] = v
	}
}

func init() {
	load(clientResources)
	load(cssResources)
	load(miscResources)
	load(jsResources)
}

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
	http.ServeContent(w, r, r.URL.Path, time.Now(), bytes.NewReader(d))
}
