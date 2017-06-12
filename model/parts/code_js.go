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
	"github.com/google/shenzhen-go/model/pin"
	"github.com/gopherjs/gopherjs/js"
)

const (
	aceGoMode      = "ace/mode/golang"
	aceJSONMode    = "ace/mode/json"
	aceChromeTheme = "ace/theme/chrome"
)

var (
	ace = jsutil.MustGetGlobal("ace")

	codePinsSession, codeImportsSession, codeHeadSession, codeBodySession, codeTailSession *js.Object

	focused *Code
)

func init() {
	codePinsSession = aceEdit("code-pins", aceJSONMode, aceChromeTheme, (*Code).handlePinsChange)
	codeImportsSession = aceEdit("code-imports", aceGoMode, aceChromeTheme, (*Code).handleImportsChange)
	codeHeadSession = aceEdit("code-head", aceGoMode, aceChromeTheme, (*Code).handleHeadChange)
	codeBodySession = aceEdit("code-body", aceGoMode, aceChromeTheme, (*Code).handleBodyChange)
	codeTailSession = aceEdit("code-tail", aceGoMode, aceChromeTheme, (*Code).handleTailChange)
}

func aceEdit(id, mode, theme string, handler func(*Code, *js.Object)) *js.Object {
	r := ace.Call("edit", id)
	if r == nil {
		log.Fatalf("Couldn't ace.edit(%q)", id)
	}
	r.Call("setTheme", theme)
	r.Set("$blockScrolling", math.Inf(1)) // Make console warnings shut up
	s := r.Call("getSession")
	s.Call("setMode", mode)
	s.Call("setUseSoftTabs", false)
	s.Call("on", "change", func(e *js.Object) {
		handler(focused, e)
	})
	return s
}

func (c *Code) handlePinsChange(e *js.Object) {
	var p []pin.Definition
	if err := json.Unmarshal([]byte(codePinsSession.Call("getValue").String()), &p); err != nil {
		// Ignore
		return
	}
	c.pins = p
}

func (c *Code) handleImportsChange(e *js.Object) {
	c.imports = strings.Split(codeImportsSession.Call("getValue").String(), "\n")
}

func (c *Code) handleHeadChange(e *js.Object) { c.head = codeHeadSession.Call("getValue").String() }
func (c *Code) handleBodyChange(e *js.Object) { c.body = codeBodySession.Call("getValue").String() }
func (c *Code) handleTailChange(e *js.Object) { c.tail = codeTailSession.Call("getValue").String() }

func (c *Code) GainFocus(*js.Object) {
	focused = c
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
