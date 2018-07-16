// Copyright 2018 Google Inc.
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

package parts

import (
	"bytes"
	"fmt"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
	"github.com/google/shenzhen-go/dev/source"
)

func init() {
	model.RegisterPartType("HTTPServeMux", "Web", &model.PartType{
		New: func() model.Part {
			return &HTTPServeMux{
				Routes: map[string]string{
					"/": "root",
				},
			}
		},
		Panels: []model.PartPanel{
			{
				Name: "Routes",
				Editor: `
				<div class="form">
					<div class="formfield">
						<input id="httpservemux-enableprometheus" name="httpservemux-enableprometheus" type="checkbox"></input>
						<label for="httpservemux-enableprometheus">Enable Prometheus metrics</label>
					</div>
				</div>
				<div class="codeedit" id="httpservemux-routes"></div>
				`,
			},
			{
				Name: "Help",
				Editor: `<div>
					<p>
						HTTPServeMux is a part which routes requests using a <code>http.ServeMux</code>.
						Refer to <a href="https://godoc.org/net/http#ServeMux">ServeMux documentation</a> for
						how ServeMux handles requests in ordinary Go.
					</p><p>
						Most requests will be forwarded to the matching output. Ordinary Go ServeMuxes
						handle some requests directly; HTTPServeMux attemps to match the same behaviour, 
						so not every input request will be sent to an output.
					</p>
				</div>`,
			},
		},
	})
}

// HTTPServeMux is a part which routes requests using a http.ServeMux.
type HTTPServeMux struct {
	EnablePrometheus bool

	// Routes is a map of patterns to output pin names.
	Routes map[string]string `json:"routes"`
}

// Clone returns a clone of this part.
func (m HTTPServeMux) Clone() model.Part {
	r := make(map[string]string, len(m.Routes))
	for k, v := range m.Routes {
		r[k] = v
	}
	return &HTTPServeMux{Routes: r}
}

// Impl returns the implementation.
func (m HTTPServeMux) Impl(string, bool, map[string]string) model.PartImpl {
	// TODO: implement prometheus metrics for this part.

	// I think http.ServeMux is concurrent safe... it guards everything with RWMutex.
	hb, tb := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	closed := source.NewStringSet()

	hb.WriteString("mux := http.NewServeMux()\n")
	for pat, out := range m.Routes {
		fmt.Fprintf(hb, "mux.Handle(%q, parts.HTTPHandler(%s))\n", pat, out)

		if closed.Ni(out) {
			continue
		}
		fmt.Fprintf(tb, "close(%s)\n", out)
		closed.Add(out)
	}

	return model.PartImpl{
		Imports: []string{
			`"net/http"`,
			`"github.com/google/shenzhen-go/dev/parts"`,
		},
		Head: hb.String(),
		// Look up the handler, assert type, and forward the original *HTTPRequest.
		// I could use mux.ServeHTTP, but this would unwrap the old *HTTPRequest
		// and wrap it in a new *HTTPRequest (that requires closing).
		// Arguably that's not an issue (channel overheads anyway), but I think it's
		// less surprising to do forwarding.
		//
		// The "*" logic was added to ServeMux's ServeHTTP to fix Go issue #3692, and
		// we should mimic the behaviour. I think it should have been implemented in
		// ServeMux.Handler; doing in ServeHTTP means anybody who is just using
		// ServeMux.Handler has to reimplement the same logic.
		//
		// http.ServeMux sometimes returns handlers defined in net/http, so handle
		// those directly.
		Body: `for req := range requests {
			// Borrow fix for Go issues #3692 and #5955.
			if req.Request.RequestURI == "*" {
				if req.Request.ProtoAtLeast(1, 1) {
					req.ResponseWriter.Header().Set("Connection", "close")
				}
				req.ResponseWriter.WriteHeader(http.StatusBadRequest)
				req.Close()
				continue
			}
			h, _ := mux.Handler(req.Request)
			hh, ok := h.(parts.HTTPHandler)
			if !ok {
				// ServeMux may return handlers that weren't added above.
				h.ServeHTTP(req.ResponseWriter, req.Request)
				req.Close()
				continue
			}
			hh <- req
		}`,
		Tail: tb.String(),
	}
}

// Pins returns a pin map, in this case varying by configuration.
func (m HTTPServeMux) Pins() pin.Map {
	p := pin.NewMap(&pin.Definition{
		Name:      "requests",
		Direction: pin.Input,
		Type:      "*parts.HTTPRequest",
	})
	for _, out := range m.Routes {
		if p[out] != nil {
			// Nothing wrong with routing multiple patterns to the same output.
			// Even if it didn't skip here, it would set the same definition...
			continue
		}
		p[out] = &pin.Definition{
			Name:      out,
			Direction: pin.Output,
			Type:      "*parts.HTTPRequest",
		}
	}
	return p
}

// TypeKey returns "HTTPServeMux".
func (m HTTPServeMux) TypeKey() string {
	return "HTTPServeMux"
}
