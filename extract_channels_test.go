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

package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestExtractChannelIdents(t *testing.T) {
	srcs, dsts, err := extractChannelIdents(`foo <- <-bar
for range baz {
    select {
    case blarp := <-qux:
    case tuz <- sax:
    }
}
close(zoop)
`)
	if err != nil {
		t.Fatalf("extractChannelIdents err = %v", err)
	}
	sort.Strings(srcs)
	sort.Strings(dsts)
	if got, want := srcs, []string{"bar", "baz", "qux"}; !reflect.DeepEqual(got, want) {
		t.Errorf("extractChannelIdents srcs = %v, want %v", got, want)
	}
	if got, want := dsts, []string{"foo", "tuz", "zoop"}; !reflect.DeepEqual(got, want) {
		t.Errorf("extractChannelIdents srcs = %v, want %v", got, want)
	}
}
