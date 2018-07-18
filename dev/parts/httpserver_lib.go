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
	"context"
	"log"
	"net/http"
	"runtime"
)

// HTTPRequest represents incoming HTTP requests and the means to respond to them.
type HTTPRequest struct {
	http.ResponseWriter
	Request *http.Request
	done    chan struct{}
}

// Close completes the request.
func (r *HTTPRequest) Close() error { close(r.done); return nil }

// HTTPHandler is a channel that can be sent HTTPRequests, thus supporting a
// simple implementation of ServerHTTP.
type HTTPHandler chan<- *HTTPRequest

func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	done := make(chan struct{})
	hr := &HTTPRequest{
		ResponseWriter: w,
		Request:        r,
		done:           done,
	}
	// Crawshaw-style sharp-edged finalizers are nice but it's possible some
	// connections will be deliberately dropped (e.g. load shedding).
	// Detect, keep calm and carry on.
	runtime.SetFinalizer(hr, func(hr *HTTPRequest) {
		log.Printf("*parts.HTTPRequest(Request: %v) not closed", r)
		close(done)
	})
	h <- hr
	<-done
}

// HTTPServerManager is information required to start and stop one HTTP server
// with the HTTPServer part.
//
// You can implement your own if you really want, but NewHTTPServerManager
// returns a simple, straightforward, channel-based implementation.
// The HTTPServer part only requires Addr and Wait.
type HTTPServerManager interface {
	// Addr is the listen address for the HTTP server.
	Addr() string

	// Shutdown passes a shutdown context to the HTTP server.
	// See https://godoc.org/net/http#Server.Shutdown for the interpretation
	// of the context.
	Shutdown(context.Context)

	// Wait waits until Shutdown is called, and then returns the context it was called with.
	Wait() context.Context
}

type httpServerManager struct {
	addr     string
	shutdown chan context.Context
}

// NewHTTPServerManager creates a channel-based HTTPServerManager.
func NewHTTPServerManager(addr string) HTTPServerManager {
	return &httpServerManager{
		addr:     addr,
		shutdown: make(chan context.Context),
	}
}

func (h *httpServerManager) Addr() string                 { return h.addr }
func (h *httpServerManager) Shutdown(ctx context.Context) { h.shutdown <- ctx }
func (h *httpServerManager) Wait() context.Context        { return <-h.shutdown }
