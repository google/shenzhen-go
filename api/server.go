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
	"net/http"
)

var (
	dispatchers = map[string]struct {
		request func() interface{}
		method  func(Interface, interface{}) (interface{}, error)
	}{
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

func dispatch(s Interface, r io.Reader) ([]byte, error) {
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

// Dispatch calls the interface function given by the request.
func Dispatch(s Interface, w http.ResponseWriter, r *http.Request) {
	rb, err := dispatch(s, r.Body)
	if err != nil {
		if st := err.(*Status); st != nil {
			w.WriteHeader(st.Code)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v", err)
		return
	}
	w.Write(rb)
}
