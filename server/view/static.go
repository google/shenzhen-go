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
	"compress/gzip"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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

type staticHandler map[string][]byte

// Static serves "static resources".
var Static = staticHandler{
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

func init() {
	Static.load(cssResources)
	Static.load(jsResources)
	Static.load(miscResources)
}

func (h staticHandler) load(m map[string][]byte) {
	for k, v := range m {
		h[k] = v
	}
}

func (h staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s static: %s", r.Method, r.URL)
	name := r.URL.Path
	d := h[name]
	if d == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	hdr := w.Header()
	hdr.Set("Cache-Control", "public")
	hdr.Set("Cache-Control", "max-age=86400")

	// Fix for LICENSE files not being detected correctly
	if strings.HasSuffix(name, "/LICENSE") {
		hdr.Set("Content-Type", "text/plain")
	}

	// Transparently handle gzipped content.
	// The content is being gzipped to reduce the binary and source code size.
	// The typical use case is running the server locally, so network is ignored.
	// So instead of on-the-fly gzipping of un-compressed content,
	// it does ..... on-the-fly un-gzipping of pre-compressed content, if gzip happens to be disallowed.

	// Content is not gzipped?
	if !bytes.Equal(d[:2], []byte{0x1f, 0x8b}) {
		//log.Printf("%s static: %s is not gzipped", r.Method, r.URL)
		http.ServeContent(w, r, name, time.Now(), bytes.NewReader(d))
		return
	}

	// Is gzip ok?
	aeh := r.Header.Get("Accept-Encoding")
	if aeh == "" {
		//log.Printf("%s static: Accept-Encoding not specified, sending gzip data", r.Method)
		hdr.Set("Content-Encoding", "gzip")
		http.ServeContent(w, r, name, time.Now(), bytes.NewReader(d))
		return
	}
	for _, ae := range strings.Split(aeh, ",") {
		bits := strings.Split(strings.TrimSpace(ae), ";")
		if bits[0] != "gzip" && bits[0] != "*" {
			continue
		}
		// If q is present, ensure q > 0.
		if len(bits) == 2 && bits[1] == "q=0" {
			// If gzip;q=0 specifically, no gzip for you.
			if bits[0] == "gzip" {
				break
			}
			continue
		}
		//log.Printf("%s static: Accept-Encoding allows gzip, sending gzip data", r.Method)
		hdr.Set("Content-Encoding", "gzip")
		http.ServeContent(w, r, name, time.Now(), bytes.NewReader(d))
		return
	}

	//log.Printf("%s static: Accept-Encoding disallows gzip, decompressing on-the-fly", r.Method)
	gr, err := gzip.NewReader(bytes.NewReader(d))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	d2, err := ioutil.ReadAll(gr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, name, time.Now(), bytes.NewReader(d2))
}
