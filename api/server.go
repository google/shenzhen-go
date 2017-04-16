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

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

var (
	dispatchers = map[string]struct {
		request func() interface{}
		method  func(Interface, interface{}) (interface{}, error)
	}{
		"SetGraphProperties": {
			request: func() interface{} { return new(SetGraphPropertiesRequest) },
			method: func(s Interface, req interface{}) (interface{}, error) {
				err := s.SetGraphProperties(req.(*SetGraphPropertiesRequest))
				return &Empty{}, err
			},
		},
		"SetNodeProperties": {
			request: func() interface{} { return new(SetNodePropertiesRequest) },
			method: func(s Interface, req interface{}) (interface{}, error) {
				err := s.SetNodeProperties(req.(*SetNodePropertiesRequest))
				return &Empty{}, err
			},
		},
		"SetPosition": {
			request: func() interface{} { return new(SetPositionRequest) },
			method: func(s Interface, req interface{}) (interface{}, error) {
				err := s.SetPosition(req.(*SetPositionRequest))
				return &Empty{}, err
			},
		},
	}
)

type jsonRequest struct {
	Method  string          `json:"method"`
	Message json.RawMessage `json:"message"`
}

func dispatch(server Interface, request io.Reader) ([]byte, error) {
	req := new(jsonRequest)
	if err := json.NewDecoder(request).Decode(req); err != nil {
		return nil, Statusf(http.StatusBadRequest, "ill-formed request: %v", err)
	}

	d, ok := dispatchers[req.Method]
	if !ok {
		return nil, Statusf(http.StatusBadRequest, "unknown method")
	}

	v := d.request()
	if err := json.Unmarshal(req.Message, v); err != nil {
		return nil, Statusf(http.StatusBadRequest, "ill-formed message: %v", err)
	}
	resp, err := d.method(server, v)
	if err != nil {
		return nil, err
	}

	w, err := json.Marshal(resp)
	if err != nil {
		return nil, Statusf(http.StatusInternalServerError, "marshalling response: %v", err)
	}
	return w, nil
}

// Dispatch calls the interface function given by the request.
func Dispatch(s Interface, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Printf("Ignoring API request with method %q", r.Method)
		return
	}

	rb, err := dispatch(s, r.Body)
	if err != nil {
		code := http.StatusInternalServerError
		if st := err.(*Status); st != nil {
			code = st.Code
		}
		w.WriteHeader(code)
		fmt.Fprintf(w, "%v", err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(rb)))
	w.Write(rb)
}
