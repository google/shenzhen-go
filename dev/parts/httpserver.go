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
)

var httpServerPins = pin.NewMap(
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
	model.RegisterPartType("HTTPServer", &model.PartType{
		New: func() model.Part { return &HTTPServer{} },
		Panels: []model.PartPanel{
			{
				Name:   "Server",
				Editor: `TODO(josh): Implement UI for editing HTTP server params`,
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
	Addr              string `json:"addr,omitempty"`
	ReadTimeout       string `json:"read_timeout,omitempty"`
	ReadHeaderTimeout string `json:"read_header_timeout,omitempty"`
	WriteTimeout      string `json:"write_timeout,omitempty"`
	IdleTimeout       string `json:"idle_timeout,omitempty"`
	MaxHeaderBytes    string `json:"max_header_bytes,omitempty"`
}

// Clone returns a clone of this HTTPServer.
func (s *HTTPServer) Clone() model.Part { s0 := *s; return &s0 }

// Impl returns the HTTPServer implementation.
func (s *HTTPServer) Impl(map[string]string) (head, body, tail string) {
	b := bytes.NewBuffer(nil)
	b.WriteString(`svr := &http.Server{
		Handler: parts.HTTPHandler(requests),
		`)
	if s.Addr != "" {
		fmt.Fprintf(b, "Addr: %s,\n", s.Addr)
	}
	if s.ReadTimeout != "" {
		fmt.Fprintf(b, "ReadTimeout: %s,\n", s.ReadTimeout)
	}
	if s.ReadHeaderTimeout != "" {
		fmt.Fprintf(b, "ReadHeaderTimeout: %s,\n", s.ReadHeaderTimeout)
	}
	if s.WriteTimeout != "" {
		fmt.Fprintf(b, "WriteTimeout: %s,\n", s.WriteTimeout)
	}
	if s.IdleTimeout != "" {
		fmt.Fprintf(b, "IdleTimeout: %s,\n", s.IdleTimeout)
	}
	if s.MaxHeaderBytes != "" {
		fmt.Fprintf(b, "MaxHeaderBytes: %s,\n", s.MaxHeaderBytes)
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

// Imports returns nil.
func (s *HTTPServer) Imports() []string {
	return []string{
		`"context"`,
		`"net/http"`,
		`"github.com/google/shenzhen-go/dev/parts"`,
	}
}

// Pins returns a map declaring a single output of any type.
func (s *HTTPServer) Pins() pin.Map { return httpServerPins }

// TypeKey returns "HTTPServer".
func (s *HTTPServer) TypeKey() string { return "HTTPServer" }
