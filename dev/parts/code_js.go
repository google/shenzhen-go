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
	"strings"

	"github.com/google/shenzhen-go/dev/dom"
	"github.com/google/shenzhen-go/dev/model/pin"
)

var (
	ace = dom.GlobalAce()

	pinsSession, importsSession, headSession, bodySession, tailSession *dom.AceSession

	focusedCode *Code
)

// Needed to resolve initialization cycle. handleFoo uses the value loaded here.
func init() {
	pinsSession = setupAce("code-pins", dom.AceJSONMode, pinsChange)
	importsSession = setupAce("code-imports", dom.AceGoMode, importsChange)
	headSession = setupAce("code-head", dom.AceGoMode, headChange)
	bodySession = setupAce("code-body", dom.AceGoMode, bodyChange)
	tailSession = setupAce("code-tail", dom.AceGoMode, tailChange)
}

func setupAce(id, mode string, handler func(dom.Object)) *dom.AceSession {
	e := ace.Edit(id)
	if e == nil {
		log.Fatalf("Couldn't ace.edit(%q)", id)
	}
	e.SetTheme(dom.AceChromeTheme)
	return e.Session().
		SetMode(mode).
		SetUseSoftTabs(false).
		On("change", handler)
}

func pinsChange(dom.Object) {
	var p pin.Map
	if err := json.Unmarshal([]byte(pinsSession.Value()), &p); err != nil {
		// Ignore
		return
	}
	focusedCode.pins = p
}

func importsChange(dom.Object) { focusedCode.imports = strings.Split(importsSession.Value(), "\n") }
func headChange(dom.Object)    { focusedCode.head = headSession.Value() }
func bodyChange(dom.Object)    { focusedCode.body = bodySession.Value() }
func tailChange(dom.Object)    { focusedCode.tail = tailSession.Value() }

func (c *Code) GainFocus() {
	focusedCode = c
	p, err := json.MarshalIndent(c.pins, "", "\t")
	if err != nil {
		// Should have parsed correctly beforehand.
		panic(err)
	}
	pinsSession.SetValue(string(p))
	importsSession.SetValue(strings.Join(c.imports, "\n"))
	headSession.SetValue(c.head)
	bodySession.SetValue(c.body)
	tailSession.SetValue(c.tail)
}
