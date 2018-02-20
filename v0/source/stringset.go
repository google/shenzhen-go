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

package source

import (
	"bytes"
	"sort"
)

// StringSet stores an unordered set of strings.
type StringSet map[string]struct{}

// NewStringSet creates a StringSet with some elements.
func NewStringSet(values ...string) StringSet {
	s := make(StringSet, len(values))
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// Union returns a set containing all elements in any of the input sets.
func Union(sets ...StringSet) StringSet {
	s := make(StringSet)
	for _, t := range sets {
		for k := range t {
			s.Add(k)
		}
	}
	return s
}

// Add adds an element to a StringSet
func (s StringSet) Add(x string) { s[x] = struct{}{} }

// Del removes an element from a StringSet
func (s StringSet) Del(x string) { delete(s, x) }

// Ni checks if an element is in the set.
func (s StringSet) Ni(x string) bool { _, y := s[x]; return y }

// Slice returns the elements in a slice, sorted.
func (s StringSet) Slice() []string {
	t := make([]string, 0, len(s))
	for k := range s {
		t = append(t, k)
	}
	return t
}

func (s StringSet) String() string {
	buf := bytes.NewBufferString("{")
	t := s.Slice()
	sort.Strings(t)
	for _, k := range t {
		buf.WriteByte(' ')
		buf.WriteString(k)
	}
	buf.WriteString(" }")
	return buf.String()
}
