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
	"testing"
)

func TestRenameChannel(t *testing.T) {
	// Nonsense program, but has shadowing identifiers
	snippet := `foo <- <-bar
<-foo
for range foo {
	select {
	case foo := <-foo:
		<-(foo.(chan interface{}))
	case art <- os.foo:
	}
}
close(foo)
close(foop)
`
	got, err := RenameChannel(snippet, "test", "foo", "quuuux")
	if err != nil {
		t.Fatalf("RenameChannel(%q, %q) error = %v", "foo", "quuuux", err)
	}
	want := `quuuux <- <-bar
<-quuuux
for range quuuux {
	select {
	case foo := <-quuuux:
		<-(foo.(chan interface{}))
	case art <- os.foo:
	}
}
close(quuuux)
close(foop)
`
	if got != want {
		t.Errorf("RenameChannel(%q, %q) = \n%q, want \n%q", "foo", "quuuux", got, want)
	}
}
