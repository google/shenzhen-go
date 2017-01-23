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

func TestRenameIdent(t *testing.T) {
	got, err := RenameIdent(`foo <- <-bar
for range bar {
    select {
    case foo := <-foo:
    case art <- os.foo:
    }
}
close(foop)
`, "test", "foo", "quuuux")
	if err != nil {
		t.Fatalf("RenameIdent(%q, %q) error = %v", "foo", "quuuux", err)
	}
	want := "quuuux <- <-bar\nfor range bar {\n\tselect {\n\tcase quuuux := <-quuuux:\n\tcase art <- os.foo:\n\t}\n}\nclose(foop)"
	if got != want {
		t.Errorf("RenameIdent(%q, %q) = %q, want %q", "foo", "quuuux", got, want)
	}
}
