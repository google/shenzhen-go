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

// ACE editor modes and themes.
const (
	ACEGoMode   = "ace/mode/golang"
	ACEJSONMode = "ace/mode/json"

	ACEChromeTheme = "ace/theme/chrome"
)

// ACE wraps an "ace" object (usually global).
type ACE struct {
	Object
}

// GlobalACE returns the global "ace" object.
func GlobalACE() *ACE {
	o := js.Global.Get("ace")
	if o == nil {
		return nil
	}
	return &ACE{WrapObject(o)}
}

// ACEEditor is an ACE editor.
type ACEEditor struct {
	Object
}

// Edit attaches an ACE edit session to an element and returns the editor object,
// or nil (if ace.edit returns null).
func (ace ACE) Edit(id string) *ACEEditor {
	o := ace.Call("edit", id)
	if o == nil {
		return nil
	}
	e := &ACEEditor{o}
	e.Set("$blockScrolling", math.Inf(1)) // Make console warnings shut up
	return e
}

// SetTheme sets the editor theme.
func (e *ACEEditor) SetTheme(theme string) *ACEEditor {
	e.Call("setTheme", theme)
	return e
}

// ACESession is an ACE editor session.
type ACESession struct {
	Object
}

// Session returns a session for this editor.
func (e *ACEEditor) Session() *ACESession {
	return &ACESession{e.Call("getSession")}
}

// SetMode sets the session mode (language).
func (s *ACESession) SetMode(mode string) *ACESession {
	s.Call("setMode", mode)
	return s
}

// SetUseSoftTabs changes the soft-tabs mode of the session.
func (s *ACESession) SetUseSoftTabs(b bool) *ACESession {
	s.Call("setUseSoftTabs", b)
	return s
}

// On adds a handler (on change, etc).
func (s *ACESession) On(event string, h func(e Object)) *ACESession {
	s.Call("on", event, func(e *js.Object) {
		h(WrapObject(e))
	})
	return s
}

// SetValue puts new contents in the session.
func (s *ACESession) SetValue(contents string) {
	s.Call("setValue", contents)
}

// Value returns the session's current contents.
func (s *ACESession) Value() string {
	return s.Call("getValue").String()
}
