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

import "net/http"

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
	h <- &HTTPRequest{
		ResponseWriter: w,
		Request:        r,
		done:           done,
	}
	<-done
}
