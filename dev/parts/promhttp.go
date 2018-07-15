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
	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

var promHTTPHandlerPins = pin.NewMap(&pin.Definition{
	Name:      "requests",
	Direction: pin.Input,
	Type:      "*parts.HTTPRequest",
})

func init() {
	model.RegisterPartType("PrometheusHTTPHandler", "Web", &model.PartType{
		New: func() model.Part { return &PrometheusHTTPHandler{} },
		Panels: []model.PartPanel{{
			Name: "Help",
			Editor: `<div>
			<p>
				A PrometheusHTTPHandler part handles requests with the Prometheus Go client.
			</p><p>
				The <code>promhttp</code> handler will serve any request this part
				receives, but Prometheus only expects the URL path to be <code>/metrics</code>.
			</p>
			</div>`,
		}},
	})
}

// PrometheusHTTPHandler is a part which immediately closes the output channel.
type PrometheusHTTPHandler struct{}

// Clone returns a clone of this PrometheusHTTPHandler.
func (PrometheusHTTPHandler) Clone() model.Part { return &PrometheusHTTPHandler{} }

// Impl returns the PrometheusHTTPHandler implementation.
func (PrometheusHTTPHandler) Impl(map[string]string) model.PartImpl {
	return model.PartImpl{
		Imports: []string{
			`"github.com/google/shenzhen-go/dev/parts"`,
			`"github.com/prometheus/client_golang/prometheus/promhttp"`,
		},
		Body: `h := promhttp.Handler()
		for r := range requests {
			h.ServeHTTP(r.ResponseWriter, r.Request)
			r.Close()
		}`,
	}
}

// Pins returns a map declaring a single request input.
func (PrometheusHTTPHandler) Pins() pin.Map { return promHTTPHandlerPins }

// TypeKey returns "PrometheusHTTPHandler".
func (PrometheusHTTPHandler) TypeKey() string { return "PrometheusHTTPHandler" }
