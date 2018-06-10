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

	codePinsSession, codeImportsSession, codeHeadSession, codeBodySession, codeTailSession *dom.ACESession

	focused *Code
)

// Needed to resolve initialization cycle. (*Code).handleFoo uses the value loaded here.
func init() {
	codePinsSession = setupACE("code-pins", dom.ACEJSONMode, (*Code).handlePinsChange)
	codeImportsSession = setupACE("code-imports", dom.ACEGoMode, (*Code).handleImportsChange)
	codeHeadSession = setupACE("code-head", dom.ACEGoMode, (*Code).handleHeadChange)
	codeBodySession = setupACE("code-body", dom.ACEGoMode, (*Code).handleBodyChange)
	codeTailSession = setupACE("code-tail", dom.ACEGoMode, (*Code).handleTailChange)
}

func setupACE(id, mode string, handler func(*Code)) *dom.ACESession {
	e := ace.Edit(id)
	if e == nil {
		log.Fatalf("Couldn't ace.edit(%q)", id)
	}
	e.SetTheme(dom.ACEChromeTheme)
	return e.Session().
		SetMode(mode).
		SetUseSoftTabs(false).
		On("change", func(dom.Object) {
			handler(focused)
		})
}

func (c *Code) handlePinsChange() {
	var p pin.Map
	if err := json.Unmarshal([]byte(codePinsSession.Value()), &p); err != nil {
		// Ignore
		return
	}
	c.pins = p
}

func (c *Code) handleImportsChange() {
	c.imports = strings.Split(codeImportsSession.Value(), "\n")
}

func (c *Code) handleHeadChange() { c.head = codeHeadSession.Value() }
func (c *Code) handleBodyChange() { c.body = codeBodySession.Value() }
func (c *Code) handleTailChange() { c.tail = codeTailSession.Value() }

func (c *Code) GainFocus() {
	focused = c
	p, err := json.MarshalIndent(c.pins, "", "\t")
	if err != nil {
		// Should have parsed correctly beforehand.
		panic(err)
	}
	codePinsSession.SetValue(string(p))
	codeImportsSession.SetValue(strings.Join(c.imports, "\n"))
	codeHeadSession.SetValue(c.head)
	codeBodySession.SetValue(c.body)
	codeTailSession.SetValue(c.tail)
}
