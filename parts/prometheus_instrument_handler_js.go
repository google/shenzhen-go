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

//+build js

package parts

import "github.com/google/shenzhen-go/dom"

var (
	selectPromInstHandlerInstrumenter = doc.ElementByID("prometheusinstrumenthandler_instrumenter")
	// TODO: implement buckets editor
	inputPromInstHandlerLabelCode   = doc.ElementByID("prometheusinstrumenthandler_labelcode")
	inputPromInstHandlerLabelMethod = doc.ElementByID("prometheusinstrumenthandler_labelmethod")

	focusedPromInstHandler *PrometheusInstrumentHandler
)

func init() {
	selectPromInstHandlerInstrumenter.AddEventListener("change", func(dom.Object) {
		focusedPromInstHandler.Instrumenter = PrometheusInstrumenter(selectPromInstHandlerInstrumenter.Get("value").String())
	})
	inputPromInstHandlerLabelCode.AddEventListener("change", func(dom.Object) {
		focusedPromInstHandler.LabelCode = inputPromInstHandlerLabelCode.Get("checked").Bool()
	})
	inputPromInstHandlerLabelMethod.AddEventListener("change", func(dom.Object) {
		focusedPromInstHandler.LabelMethod = inputPromInstHandlerLabelMethod.Get("checked").Bool()
	})
}

func (h *PrometheusInstrumentHandler) GainFocus() {
	focusedPromInstHandler = h
	selectPromInstHandlerInstrumenter.Set("value", h.Instrumenter)
	inputPromInstHandlerLabelCode.Set("value", h.LabelCode)
	inputPromInstHandlerLabelMethod.Set("value", h.LabelMethod)
}
