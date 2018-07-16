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
	model.RegisterPartType("PrometheusMetricsHandler", "Web", &model.PartType{
		New: func() model.Part { return &PrometheusMetricsHandler{} },
		Panels: []model.PartPanel{{
			Name: "Help",
			Editor: `<div>
			<p>
				A PrometheusMetricsHandler part handles requests with the Prometheus Go client:
				it responds by using the handler returned by <code>promhttp.Handler</code>.
			</p><p>
				This part will serve any request it	receives with <code>promhttp.Handler</code>,
				but Prometheus generally expects the URL path to be <code>/metrics</code>.
			</p>
			</div>`,
		}},
	})
}

// PrometheusMetricsHandler is a part which immediately closes the output channel.
type PrometheusMetricsHandler struct{}

// Clone returns a clone of this PrometheusMetricsHandler.
func (PrometheusMetricsHandler) Clone() model.Part { return &PrometheusMetricsHandler{} }

// Impl returns the PrometheusMetricsHandler implementation.
func (PrometheusMetricsHandler) Impl(string, bool, map[string]string) model.PartImpl {
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
func (PrometheusMetricsHandler) Pins() pin.Map { return promHTTPHandlerPins }

// TypeKey returns "PrometheusMetricsHandler".
func (PrometheusMetricsHandler) TypeKey() string { return "PrometheusMetricsHandler" }
