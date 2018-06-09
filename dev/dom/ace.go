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

package dom

import (
	"math"

	"github.com/gopherjs/gopherjs/js"
)

// ACE wraps an "ace" object (usually global).
type ACE struct {
	Object
}

// GlobalACE returns the global "ace" object.
func GlobalACE() ACE {
	return ACE{WrapObject(js.Global.Get("ace"))}
}

// ACEEditor is an ACE editor.
type ACEEditor struct {
	Object
}

// Edit attaches an ACE edit session to an element and returns the
func (ace ACE) Edit(id string) ACEEditor {
	return ACEEditor{ace.Call("edit", id)}
}

func misc(r Object, theme, mode string) Object {
	r.Call("setTheme", theme)
	r.Set("$blockScrolling", math.Inf(1)) // Make console warnings shut up
	s := r.Call("getSession")
	s.Call("setMode", mode)
	s.Call("setUseSoftTabs", false)
	s.Call("on", "change", func(e *js.Object) {
	})
	return s
}
