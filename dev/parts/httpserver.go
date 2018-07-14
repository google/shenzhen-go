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
	"time"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

var httpServerPins = pin.NewMap(
	&pin.Definition{
		Name:      "addr",
		Direction: pin.Input,
		Type:      "string",
	},
	&pin.Definition{
		Name:      "shutdown",
		Direction: pin.Input,
		Type:      "context.Context",
	},
	&pin.Definition{
		Name:      "requests",
		Direction: pin.Output,
		Type:      "*parts.HTTPRequest",
	},
	&pin.Definition{
		Name:      "errors",
		Direction: pin.Output,
		Type:      "error",
	},
)

func init() {
	model.RegisterPartType("HTTPServer", "Web", &model.PartType{
		New: func() model.Part { return &HTTPServer{} },
		Panels: []model.PartPanel{
			{
				Name: "Server",
				Editor: `<div>
				<div class="formfield">
					<label for="httpserver-readtimeout">Read timeout</label>
					<input id="httpserver-readtimeout" name="httpserver-readtimeout" type="text" required title="Must be a parseable time.Duration" value="0s"></input>
				</div>
				<div class="formfield">
					<label for="httpserver-readheadertimeout">Read header timeout</label>
					<input id="httpserver-readheadertimeout" name="httpserver-readheadertimeout" type="text" required title="Must be a parseable time.Duration" value="0s"></input>
				</div>
				<div class="formfield">
					<label for="httpserver-writetimeout">Write timeout</label>
					<input id="httpserver-writetimeout" name="httpserver-writetimeout" type="text" required title="Must be a parseable time.Duration" value="0s"></input>
				</div>
				<div class="formfield">
					<label for="httpserver-idletimeout">Idle timeout</label>
					<input id="httpserver-idletimeout" name="httpserver-idletimeout" type="text" required title="Must be a parseable time.Duration" value="0s"></input>
				</div>
				<div class="formfield">
					<label for="httpserver-maxheaderbytes">Max header bytes</label>
					<input id="httpserver-maxheaderbytes" name="httpserver-maxheaderbytes" type="number" required title="Must be a whole number. 0 means http.DefaultMaxHeaderBytes." value="0"></input>
				</div>
			</div>`,
			},
			{
				Name: "Help",
				Editor: `<div>
			<p>
				A HTTPServer part serves HTTP requests.
			</p>
			</div>`,
			},
		},
	})
}

// HTTPServer is a part which listens on an address and
// serves HTTP requests.
type HTTPServer struct {
	ReadTimeout       time.Duration `json:"read_timeout,omitempty"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout,omitempty"`
	WriteTimeout      time.Duration `json:"write_timeout,omitempty"`
	IdleTimeout       time.Duration `json:"idle_timeout,omitempty"`
	MaxHeaderBytes    int           `json:"max_header_bytes,omitempty"`
}

// Clone returns a clone of this HTTPServer.
func (s *HTTPServer) Clone() model.Part { s0 := *s; return &s0 }

// Impl returns the HTTPServer implementation.
func (s *HTTPServer) Impl(map[string]string) (head, body, tail string) {
	b := bytes.NewBuffer(nil)
	b.WriteString(`svr := &http.Server{
		Handler: parts.HTTPHandler(requests),
		Addr:    <-addr,
		`)
	if s.ReadTimeout != 0 {
		fmt.Fprintf(b, "ReadTimeout: %d, // %v\n", s.ReadTimeout, s.ReadTimeout)
	}
	if s.ReadHeaderTimeout != 0 {
		fmt.Fprintf(b, "ReadHeaderTimeout: %d, // %v\n", s.ReadHeaderTimeout, s.ReadHeaderTimeout)
	}
	if s.WriteTimeout != 0 {
		fmt.Fprintf(b, "WriteTimeout: %d, // %v\n", s.WriteTimeout, s.WriteTimeout)
	}
	if s.IdleTimeout != 0 {
		fmt.Fprintf(b, "IdleTimeout: %d, // %v\n", s.IdleTimeout, s.IdleTimeout)
	}
	if s.MaxHeaderBytes != 0 {
		fmt.Fprintf(b, "MaxHeaderBytes: %d,\n", s.MaxHeaderBytes)
	}
	b.WriteString(`}
	var shutdone chan struct{}
	go func() {
		ctx := <-shutdown
		shutdone = make(chan struct{})
		svr.Shutdown(ctx)
		close(shutdone)
	}()
	err := svr.ListenAndServe()
	if errors != nil {
		errors <- err
	}
	if shutdone != nil {
		<-shutdone 
	}`)
	return "",
		b.String(),
		`close(requests)
		if errors != nil {
			close(errors)
		}
		`
}

// Imports returns some needed imports.
func (s *HTTPServer) Imports() []string {
	return []string{
		`"context"`,
		`"net/http"`,
		`"github.com/google/shenzhen-go/dev/parts"`,
	}
}

// Pins returns a map declaring a bunch of pins.
func (s *HTTPServer) Pins() pin.Map { return httpServerPins }

// TypeKey returns "HTTPServer".
func (s *HTTPServer) TypeKey() string { return "HTTPServer" }
