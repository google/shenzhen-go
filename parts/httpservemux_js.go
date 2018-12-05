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

import (
	"encoding/json"
	"log"

	"github.com/google/shenzhen-go/dom"
)

var (
	httpServeMuxRoutesSession *dom.AceSession

	inputHTTPServeMuxEnablePrometheus = doc.ElementByID("httpservemux-enableprometheus")

	focusedHTTPServeMux *HTTPServeMux
)

func init() {
	httpServeMuxRoutesSession = setupAce("httpservemux-routes", dom.AceJSONMode, httpServeMuxRoutesChange)
	inputHTTPServeMuxEnablePrometheus.AddEventListener("change", func(dom.Object) {
		focusedHTTPServeMux.EnablePrometheus = inputHTTPServeMuxEnablePrometheus.Get("checked").Bool()
	})
}

func httpServeMuxRoutesChange(dom.Object) {
	routes := make(map[string]string)
	if err := json.Unmarshal([]byte(httpServeMuxRoutesSession.Value()), &routes); err != nil {
		log.Printf("Couldn't unmarshal httpServeMuxRoutesSession value into a map[string]string: %v", err)
		return
	}
	focusedHTTPServeMux.Routes = routes
}

func (m *HTTPServeMux) GainFocus() {
	focusedHTTPServeMux = m
	routes, err := json.MarshalIndent(m.Routes, "", "\t")
	if err != nil {
		// ... how?
		log.Fatalf("Couldn't marshal a map[string]string to JSON: %v", err)
	}
	httpServeMuxRoutesSession.SetValue(string(routes))
	inputHTTPServeMuxEnablePrometheus.Set("checked", focusedHTTPServeMux.EnablePrometheus)
}
