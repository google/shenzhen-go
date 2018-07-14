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
	codePinsSession, codeImportsSession, codeHeadSession, codeBodySession, codeTailSession *dom.AceSession

	focusedCode *Code
)

// Needed to resolve initialization cycle. handleFoo uses the value loaded here.
func init() {
	codePinsSession = setupAce("code-pins", dom.AceJSONMode, codePinsChange)
	codeImportsSession = setupAce("code-imports", dom.AceGoMode, codeImportsChange)
	codeHeadSession = setupAce("code-head", dom.AceGoMode, codeHeadChange)
	codeBodySession = setupAce("code-body", dom.AceGoMode, codeBodyChange)
	codeTailSession = setupAce("code-tail", dom.AceGoMode, codeTailChange)
}

func codePinsChange(dom.Object) {
	var p pin.Map
	if err := json.Unmarshal([]byte(codePinsSession.Value()), &p); err != nil {
		log.Printf("Couldn't unmarshal codePinsSession value into a pin.Map: %v", err)
		return
	}
	focusedCode.pins = p
}

func codeImportsChange(dom.Object) {
	focusedCode.imports = strings.Split(codeImportsSession.Value(), "\n")
}

func codeHeadChange(dom.Object) { focusedCode.head = codeHeadSession.Value() }
func codeBodyChange(dom.Object) { focusedCode.body = codeBodySession.Value() }
func codeTailChange(dom.Object) { focusedCode.tail = codeTailSession.Value() }

func (c *Code) GainFocus() {
	focusedCode = c
	p, err := json.MarshalIndent(c.pins, "", "\t")
	if err != nil {
		// ...How?
		log.Fatalf("Couldn't marshal a pin.Map to JSON: %v", err)
	}
	codePinsSession.SetValue(string(p))
	codeImportsSession.SetValue(strings.Join(c.imports, "\n"))
	codeHeadSession.SetValue(c.head)
	codeBodySession.SetValue(c.body)
	codeTailSession.SetValue(c.tail)
}
