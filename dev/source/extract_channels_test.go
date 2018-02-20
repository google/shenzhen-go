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

package source

import (
	"reflect"
	"testing"
)

func TestExtractChannels(t *testing.T) {
	srcs, dsts, err := ExtractChannels(`foo <- <-bar
for range baz {
    select {
    case blarp := <-qux:
    case tuz <- sax:
    }
}
for range other {
	// other is not a declared channel
}
close(zoop)
`, "demo", `var bar, baz, foo, tuz, qux, zoop chan interface{}`)
	if err != nil {
		t.Fatalf("ExtractChannels error = %v", err)
	}
	if got, want := srcs, NewStringSet("bar", "baz", "qux"); !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractChannels srcs = %v, want %v", got, want)
	}
	if got, want := dsts, NewStringSet("foo", "tuz", "zoop"); !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractChannels dsts = %v, want %v", got, want)
	}
}
