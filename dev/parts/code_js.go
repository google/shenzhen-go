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
	ace = dom.GlobalACE()

	pinsSession, importsSession, headSession, bodySession, tailSession *dom.ACESession

	focused *Code
)

// Needed to resolve initialization cycle. handleFoo uses the value loaded here.
func init() {
	pinsSession = setupACE("code-pins", dom.ACEJSONMode, pinsChange)
	importsSession = setupACE("code-imports", dom.ACEGoMode, importsChange)
	headSession = setupACE("code-head", dom.ACEGoMode, headChange)
	bodySession = setupACE("code-body", dom.ACEGoMode, bodyChange)
	tailSession = setupACE("code-tail", dom.ACEGoMode, tailChange)
}

func setupACE(id, mode string, handler func(dom.Object)) *dom.ACESession {
	e := ace.Edit(id)
	if e == nil {
		log.Fatalf("Couldn't ace.edit(%q)", id)
	}
	e.SetTheme(dom.ACEChromeTheme)
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
	focused.pins = p
}

func importsChange(dom.Object) { focused.imports = strings.Split(importsSession.Value(), "\n") }
func headChange(dom.Object)    { focused.head = headSession.Value() }
func bodyChange(dom.Object)    { focused.body = bodySession.Value() }
func tailChange(dom.Object)    { focused.tail = tailSession.Value() }

func (c *Code) GainFocus() {
	focused = c
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
