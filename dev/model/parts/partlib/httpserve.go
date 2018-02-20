// Copyright 2017 Google Inc.
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

package partlib

import (
	"net/http"
)

type request struct {
	http.ResponseWriter
	r    *http.Request
	done chan struct{}
}

func (r *request) Info() *http.Request { return r.r }
func (r *request) Close() error        { close(r.done); return nil }

// HTTPRequest represents incoming HTTP requests and the means to respond to them.
type HTTPRequest interface {
	http.ResponseWriter
	Info() *http.Request
	Close() error
}

// HTTPHandlerChan is a channel that can be sent HTTPRequests, thus supporting a
// simple implementation of ServerHTTP.
type HTTPHandlerChan chan<- HTTPRequest

func (h HTTPHandlerChan) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	done := make(chan struct{})
	h <- &request{
		ResponseWriter: w,
		r:              r,
		done:           done,
	}
	<-done
}
