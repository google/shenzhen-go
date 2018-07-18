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
	"fmt"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

var prometheusInstrumentHandlerPins = pin.NewMap(
	&pin.Definition{
		Name:      "in",
		Direction: pin.Input,
		Type:      "*parts.HTTPRequest",
	},
	&pin.Definition{
		Name:      "out",
		Direction: pin.Output,
		Type:      "*parts.HTTPRequest",
	},
)

func init() {
	model.RegisterPartType("PrometheusInstrumentHandler", "Web", &model.PartType{
		New: func() model.Part {
			return &PrometheusInstrumentHandler{
				Instrumenter: PromInstDuration,
				// default buckets, no code or method labels.
			}
		},
		Panels: []model.PartPanel{
			{
				Name: "Options",
				Editor: `<div class="form">
					<div class="formfield">
						<label for="prometheusinstrumenthandler_instrumenter">Instrumenter</label>
						<select id="prometheusinstrumenthandler_instrumenter" name="prometheusinstrumenthandler_instrumenter">
							<option value="Duration" selected>Duration</option>
							<option value="RequestSize">RequestSize</option>
							<option value="ResponseSize">ResponseSize</option>
							<option value="TimeToWriteHeader">TimeToWriteHeader</option>
						</select>
					</div>
					<div class="formfield">
						TODO: implement buckets
					</div>
					<div class="formfield">
						<input type="checkbox" id="prometheusinstrumenthandler_labelcode" name="prometheusinstrumenthandler_labelcode"></input>
						<label for="prometheusinstrumenthandler_labelcode">Code label</label>
					</div>
					<div class="formfield">
					<input type="checkbox" id="prometheusinstrumenthandler_labelmethod" name="prometheusinstrumenthandler_labelmethod"></input>
					<label for="prometheusinstrumenthandler_labelmethod">Method label</label>
					</div>
				</div>`,
			},
			{
				Name: "Help",
				Editor: `<div>
			<p>
				A PrometheusInstrumentHandler part wraps the handler attached to the "out"
				with instrumenting code. It uses the <code>promhttp.InstrumentHandler${X}</code>
				series of functions (where "${X}" is one of Duration, RequestSize, ResponseSize, 
				or TimeToWriteHeader).
			</p><p>
				TODO: implement InstrumentHandlerCounter (counter) and InstrumentHandlerInFlight
				(gauge). The implemented ones are all histograms.
			</p>
			</div>`,
			},
		},
	})
}

// PrometheusInstrumentHandler is a part which immediately closes the output channel.
type PrometheusInstrumentHandler struct {
	Instrumenter PrometheusInstrumenter `json:"instrumenter"`
	Buckets      []float64              `json:"buckets,omitempty"`
	LabelCode    bool                   `json:"label_code"`
	LabelMethod  bool                   `json:"label_method"`
}

// PrometheusInstrumenter specifies one of the Prometheus instrument-handlers.
type PrometheusInstrumenter string

// Available Prometheus instrumenters.
const (
	//PromInstCounter           PrometheusInstrumenter = "Counter"
	PromInstDuration PrometheusInstrumenter = "Duration"
	//PromInstInFlight          PrometheusInstrumenter = "InFlight"
	PromInstRequestSize       PrometheusInstrumenter = "RequestSize"
	PromInstResponseSize      PrometheusInstrumenter = "ResponseSize"
	PromInstTimeToWriteHeader PrometheusInstrumenter = "TimeToWriteHeader"
)

func (i PrometheusInstrumenter) help() string {
	switch i {
	case PromInstDuration:
		return "Durations of requests"
	case PromInstRequestSize:
		return "Sizes of requests"
	case PromInstResponseSize:
		return "Sizes of responses"
	case PromInstTimeToWriteHeader:
		return "Times to write response header"
	default:
		panic("unsupported instrumenter " + i)
	}
}
func (h *PrometheusInstrumentHandler) labels() []string {
	var s []string
	if h.LabelCode {
		s = append(s, "code")
	}
	if h.LabelMethod {
		s = append(s, "method")
	}
	return s
}

// Clone returns a clone of this PrometheusInstrumentHandler.
func (h *PrometheusInstrumentHandler) Clone() model.Part {
	h0 := *h
	return &h0
}

// Impl returns the PrometheusInstrumentHandler implementation.
func (h *PrometheusInstrumentHandler) Impl(n *model.Node) model.PartImpl {
	return model.PartImpl{
		Imports: []string{
			`"github.com/google/shenzhen-go/dev/parts"`,
			`"github.com/prometheus/client_golang/prometheus"`,
			`"github.com/prometheus/client_golang/prometheus/promhttp"`,
		},
		// Because the buckets can vary by node, it needs a different metric per part. (argh!)
		Head: fmt.Sprintf(`
		sum := prometheus.NewHistogramVec(
					prometheus.HistogramOpts{
						Namespace: "shenzhen_go",
						Subsystem: "instrument_handler",
						Name:      %q,
						Help:      %q,
						Buckets:   %#v,
					},
					%#v)
		prometheus.MustRegister(sum)
		`, model.Mangle(n.Name), h.Instrumenter.help(), h.Buckets, h.labels()),
		Body: fmt.Sprintf(`
		h := promhttp.InstrumentHandler%s(sum, parts.HTTPHandler(out))
		for r := range in {
			h.ServeHTTP(r.ResponseWriter, r.Request)
			r.Close()
		}`, h.Instrumenter),
		Tail: `close(out)`,
	}
}

// Pins returns a map declaring a single request input.
func (h *PrometheusInstrumentHandler) Pins() pin.Map { return prometheusInstrumentHandlerPins }

// TypeKey returns "PrometheusInstrumentHandler".
func (h *PrometheusInstrumentHandler) TypeKey() string { return "PrometheusInstrumentHandler" }
