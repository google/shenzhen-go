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

//+build js

package parts

import (
	"encoding/json"
	"log"
	"math"
	"strings"

	"github.com/google/shenzhen-go/jsutil"
	"github.com/gopherjs/gopherjs/js"
)

const (
	aceGoMode      = "ace/mode/golang"
	aceJSONMode    = "ace/mode/json"
	aceChromeTheme = "ace/theme/chrome"
)

var (
	ace = jsutil.MustGetGlobal("ace")

	_, codePinsSession    = aceEdit("code-pins", aceJSONMode, aceChromeTheme)
	_, codeImportsSession = aceEdit("code-imports", aceGoMode, aceChromeTheme)
	_, codeHeadSession    = aceEdit("code-head", aceGoMode, aceChromeTheme)
	_, codeBodySession    = aceEdit("code-body", aceGoMode, aceChromeTheme)
	_, codeTailSession    = aceEdit("code-tail", aceGoMode, aceChromeTheme)
)

func aceEdit(id, mode, theme string) (editor, session *js.Object) {
	r := ace.Call("edit", id)
	if r == nil {
		log.Fatalf("Couldn't ace.edit(%q)", id)
	}
	r.Call("setTheme", theme)
	r.Set("$blockScrolling", math.Inf(1)) // Make console warnings shut up
	s := r.Call("getSession")
	s.Call("setMode", mode)
	s.Call("setUseSoftTabs", false)
	return r, s
}

func (c *Code) GainFocus(*js.Object) {
	p, err := json.MarshalIndent(c.pins, "", "\t")
	if err != nil {
		// Should have parsed correctly beforehand.
		panic(err)
	}
	codePinsSession.Call("setValue", string(p))
	codeImportsSession.Call("setValue", strings.Join(c.imports, "\n"))
	codeHeadSession.Call("setValue", c.head)
	codeBodySession.Call("setValue", c.body)
	codeTailSession.Call("setValue", c.tail)
}
