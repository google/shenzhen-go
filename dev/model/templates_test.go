// Copyright 2016 Google Inc.
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

package model

import "testing"

type nopWriter struct{}

func (nopWriter) Write(r []byte) (int, error) { return len(r), nil }

func TestGoTemplate(t *testing.T) {
	// Smoke-testing the template.
	for name, g := range TestGraphs {
		if err := goTemplate.Execute(nopWriter{}, g); err != nil {
			t.Errorf("goTemplate.Execute(%v) = %v, want nil error", name, err)
		}
	}
}
