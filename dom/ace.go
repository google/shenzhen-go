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
	"syscall/js"
)

// Ace editor modes and themes.
const (
	// Ace modes.
	AceGoMode   = "ace/mode/golang"
	AceJSONMode = "ace/mode/json"

	// Ace themes.
	AceChromeTheme              = "ace/theme/chrome"
	AceTomorrowNightBrightTheme = "ace/theme/tomorrow_night_bright"
)

// Ace wraps an "ace" object (usually global).
type Ace struct {
	js.Value
}

// GlobalAce returns the global "ace" object.
func GlobalAce() Ace {
	return Ace{Value: js.Global().Get("ace")}
}

// AceEditor is an Ace editor.
type AceEditor struct {
	js.Value
}

// Edit attaches an Ace edit session to an element and returns the editor object,
// or nil (if ace.edit returns null).
func (ace Ace) Edit(id string) AceEditor {
	e := AceEditor{Value: ace.Call("edit", id)}
	e.Set("$blockScrolling", math.Inf(1)) // Make console warnings shut up
	return e
}

// SetTheme sets the editor theme.
func (e AceEditor) SetTheme(theme string) AceEditor {
	e.Call("setTheme", theme)
	return e
}

// AceSession is an Ace editor session.
type AceSession struct {
	js.Value
}

// Session returns a session for this editor.
func (e AceEditor) Session() AceSession {
	return AceSession{Value: e.Call("getSession")}
}

// SetMode sets the session mode (language).
func (s AceSession) SetMode(mode string) AceSession {
	s.Call("setMode", mode)
	return s
}

// SetUseSoftTabs changes the soft-tabs mode of the session.
func (s AceSession) SetUseSoftTabs(b bool) AceSession {
	s.Call("setUseSoftTabs", b)
	return s
}

// On adds a handler (on change, etc).
func (s AceSession) On(event string, h js.Callback) AceSession {
	s.Call("on", event, h)
	return s
}

// SetContents puts new contents in the session.
func (s AceSession) SetContents(contents string) {
	s.Call("setValue", contents)
}

// Contents returns the session's current contents.
func (s AceSession) Contents() string {
	return s.Call("getValue").String()
}
