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
	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

func init() {
	model.RegisterPartType("HTTPServeMux", "Web", &model.PartType{
		New: func() model.Part { return &HTTPServeMux{} },
		Panels: []model.PartPanel{{
			Name:   "Help",
			Editor: `<div><p>HTTPServeMux is a part which routes requests using a <code>http.ServeMux</code>.</p></div>`,
		}},
	})
}

// HTTPServeMux is a part which routes requests using a http.ServeMux.
type HTTPServeMux struct {
	// TODO(josh): Implement.
}

// Clone returns a clone of this part.
func (m *HTTPServeMux) Clone() model.Part {
	m0 := *m
	return &m0
}

// Impl returns the implementation.
func (m *HTTPServeMux) Impl(types map[string]string) (head, body, tail string) {
	// TODO(josh): Implement.
	return "", "", ""
}

// Imports returns needed imports.
func (m *HTTPServeMux) Imports() []string {
	return []string{
		`"net/http"`,
		`"github.com/google/shenzhen-go/dev/parts"`,
	}
}

// Pins returns a pin map, in this case varying by configuration.
func (m *HTTPServeMux) Pins() pin.Map {
	// TODO(josh): Implement.
	return nil
}

// TypeKey returns "HTTPServeMux".
func (m *HTTPServeMux) TypeKey() string {
	return "HTTPServeMux"
}
