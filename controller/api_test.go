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

package controller

import (
	"testing"

	"github.com/google/shenzhen-go/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func code(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	if st, ok := status.FromError(err); ok {
		return st.Code()
	}
	return codes.Unknown
}

func TestLookupGraph(t *testing.T) {
	c := &controller{
		loadedGraphs: map[string]*model.Graph{
			"foo": &model.Graph{},
			"bar": &model.Graph{},
		},
	}
	tests := []struct {
		key  string
		code codes.Code
	}{
		{"foo", codes.OK},
		{"bar", codes.OK},
		{"baz", codes.NotFound},
	}
	for _, test := range tests {
		_, err := c.lookupGraph(test.key)
		if got, want := code(err), test.code; got != want {
			t.Errorf("c.lookupGraph(%q) = %v; want %v", test.key, got, want)
		}
	}
}
