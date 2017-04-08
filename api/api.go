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

// Package api has types for communicating with the UI.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	dispatchers = map[string]struct {
		request func() interface{}
		method  func(Server, interface{}) (interface{}, error)
	}{
		"SetPosition": {
			request: func() interface{} { return new(SetPositionRequest) },
			method: func(s Server, req interface{}) (interface{}, error) {
				err := s.SetPosition(req.(*SetPositionRequest))
				return &Empty{}, err
			},
		},
	}
)

// Status bundles a HTTP status code together with an error reason.
type Status struct {
	Code   int
	Reason error
}

func (s *Status) Error() string {
	return fmt.Sprintf("status %d: %v", s.Code, s.Reason)
}

// Statusf creates a Status with a format string and arguments.
func Statusf(code int, f string, args ...interface{}) *Status {
	if code == http.StatusOK {
		return nil
	}
	return &Status{
		Code:   code,
		Reason: fmt.Errorf(f, args...),
	}
}

// Server is the interface that is used to handle requests.
type Server interface {
	SetPosition(*SetPositionRequest) error
}

type jsonRequest struct {
	Method  string          `json:"method"`
	Message json.RawMessage `json:"message"`
}

// Dispatch calls the server function given by the request.
func Dispatch(s Server, r io.Reader) ([]byte, error) {
	var req jsonRequest
	if err := json.NewDecoder(r).Decode(&req); err != nil {
		return nil, Statusf(http.StatusBadRequest, "ill-formed request: %v")
	}

	d, ok := dispatchers[req.Method]
	if !ok {
		return nil, Statusf(http.StatusBadRequest, "unknown method")
	}

	v := d.request()
	if err := json.Unmarshal(req.Message, v); err != nil {
		return nil, Statusf(http.StatusBadRequest, "ill-formed message: %v", err)
	}
	resp, err := d.method(s, v)
	if err != nil {
		return nil, err
	}

	w, err := json.Marshal(resp)
	if err != nil {
		return nil, Statusf(http.StatusInternalServerError, "marshalling response: %v", err)
	}
	return w, nil
}

// Empty is just an empty message.
type Empty struct{}

// Request is the embedded base of all requests.
type Request struct {
	Graph string `json:"graph"`
}

// ChannelRequest is the embedded base of all requests to do with channels.
type ChannelRequest struct {
	Request
	Channel string `json:"channel"`
}

// NodeRequest is the embedded base of all requests to do with nodes.
type NodeRequest struct {
	Request
	Node string `json:"node"`
}

// SetPositionRequest is a request to change the position of a node.
type SetPositionRequest struct {
	NodeRequest
	X int `json:"x"`
	Y int `json:"y"`
}
