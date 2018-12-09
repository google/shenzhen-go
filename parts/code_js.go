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
	"syscall/js"

	"github.com/google/shenzhen-go/dom"
	"github.com/google/shenzhen-go/model/pin"
)

var (
	codePinsSession, codeImportsSession, codeHeadSession, codeBodySession, codeTailSession dom.AceSession

	linkCodeFormatHead = doc.ElementByID("code-format-head-link")
	linkCodeFormatBody = doc.ElementByID("code-format-body-link")
	linkCodeFormatTail = doc.ElementByID("code-format-tail-link")

	focusedCode *Code
)

// Needed to resolve initialization cycle. handleFoo uses the value loaded here.
func init() {
	codePinsSession = setupAce("code-pins", dom.AceJSONMode, js.NewEventCallback(0, codePinsChange))
	codeImportsSession = setupAce("code-imports", dom.AceGoMode, js.NewEventCallback(0, codeImportsChange))
	codeHeadSession = setupAce("code-head", dom.AceGoMode, js.NewEventCallback(0, codeHeadChange))
	codeBodySession = setupAce("code-body", dom.AceGoMode, js.NewEventCallback(0, codeBodyChange))
	codeTailSession = setupAce("code-tail", dom.AceGoMode, js.NewEventCallback(0, codeTailChange))

	linkCodeFormatHead.AddEventListener("click", formatHandler(codeHeadSession))
	linkCodeFormatBody.AddEventListener("click", formatHandler(codeBodySession))
	linkCodeFormatTail.AddEventListener("click", formatHandler(codeTailSession))
}

func codePinsChange(js.Value) {
	var p pin.Map
	if err := json.Unmarshal([]byte(codePinsSession.Contents()), &p); err != nil {
		log.Printf("Couldn't unmarshal codePinsSession value into a pin.Map: %v", err)
		return
	}
	focusedCode.PinMap = p
}

func codeImportsChange(js.Value) {
	focusedCode.Imports = stripCR(strings.Split(codeImportsSession.Contents(), "\n"))
}

func codeHeadChange(js.Value) {
	focusedCode.Head = stripCR(strings.Split(codeHeadSession.Contents(), "\n"))
}

func codeBodyChange(js.Value) {
	focusedCode.Body = stripCR(strings.Split(codeBodySession.Contents(), "\n"))
}

func codeTailChange(js.Value) {
	focusedCode.Tail = stripCR(strings.Split(codeTailSession.Contents(), "\n"))
}

func (c *Code) GainFocus() {
	focusedCode = c
	p, err := json.MarshalIndent(c.PinMap, "", "\t")
	if err != nil {
		// ...How?
		log.Fatalf("Couldn't marshal a pin.Map to JSON: %v", err)
	}
	codePinsSession.SetContents(string(p))
	codeImportsSession.SetContents(strings.Join(c.Imports, "\n"))
	codeHeadSession.SetContents(strings.Join(c.Head, "\n"))
	codeBodySession.SetContents(strings.Join(c.Body, "\n"))
	codeTailSession.SetContents(strings.Join(c.Tail, "\n"))
}
