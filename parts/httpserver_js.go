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
	"log"
	"time"

	"github.com/google/shenzhen-go/dom"
)

var (
	httpServerOutlets = struct {
		inputReadTimeout       dom.Element
		inputReadHeaderTimeout dom.Element
		inputWriteTimeout      dom.Element
		inputIdleTimeout       dom.Element
		inputMaxHeaderBytes    dom.Element
	}{
		inputReadTimeout:       doc.ElementByID("httpserver-readtimeout"),
		inputReadHeaderTimeout: doc.ElementByID("httpserver-readheadertimeout"),
		inputWriteTimeout:      doc.ElementByID("httpserver-writetimeout"),
		inputIdleTimeout:       doc.ElementByID("httpserver-idletimeout"),
		inputMaxHeaderBytes:    doc.ElementByID("httpserver-maxheaderbytes"),
	}

	focusedHTTPServer *HTTPServer
)

func durationChange(change func(time.Duration)) func(dom.Object) {
	return func(ev dom.Object) {
		in := ev.Get("target").Get("value").String()
		t, err := time.ParseDuration(in)
		if err != nil {
			log.Printf("value %q is not a time.Duration", in)
			return
		}
		change(t)
	}
}

func init() {
	httpServerOutlets.inputReadTimeout.AddEventListener("change", durationChange(setReadTimeout))
	httpServerOutlets.inputReadHeaderTimeout.AddEventListener("change", durationChange(setReadHeaderTimeout))
	httpServerOutlets.inputWriteTimeout.AddEventListener("change", durationChange(setWriteTimeout))
	httpServerOutlets.inputIdleTimeout.AddEventListener("change", durationChange(setIdleTimeout))
	httpServerOutlets.inputMaxHeaderBytes.AddEventListener("change", func(dom.Object) {
		focusedHTTPServer.MaxHeaderBytes = httpServerOutlets.inputMaxHeaderBytes.Get("value").Int()
	})
}

func setReadTimeout(t time.Duration)       { focusedHTTPServer.ReadTimeout = t }
func setReadHeaderTimeout(t time.Duration) { focusedHTTPServer.ReadHeaderTimeout = t }
func setWriteTimeout(t time.Duration)      { focusedHTTPServer.WriteTimeout = t }
func setIdleTimeout(t time.Duration)       { focusedHTTPServer.IdleTimeout = t }

func (s *HTTPServer) GainFocus() {
	focusedHTTPServer = s
	httpServerOutlets.inputReadTimeout.Set("value", focusedHTTPServer.ReadTimeout.String())
	httpServerOutlets.inputReadHeaderTimeout.Set("value", focusedHTTPServer.ReadHeaderTimeout.String())
	httpServerOutlets.inputWriteTimeout.Set("value", focusedHTTPServer.WriteTimeout.String())
	httpServerOutlets.inputIdleTimeout.Set("value", focusedHTTPServer.IdleTimeout.String())
	httpServerOutlets.inputMaxHeaderBytes.Set("value", focusedHTTPServer.MaxHeaderBytes)
}
