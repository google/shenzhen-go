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

//+build js

package parts

import (
	"strings"

	"github.com/google/shenzhen-go/dom"
)

var (
	transformImportsSession, transformBodySession *dom.AceSession

	inputTransformInputType  = doc.ElementByID("transform-inputtype")
	inputTransformOutputType = doc.ElementByID("transform-outputtype")
	linkTransformFormat      = doc.ElementByID("transform-format-link")

	focusedTransform *Transform
)

// Needed to resolve initialization cycle. handleFoo uses the value loaded here.
func init() {
	transformImportsSession = setupAce("transform-imports", dom.AceGoMode, transformImportsChange)
	transformBodySession = setupAce("transform-body", dom.AceGoMode, transformBodyChange)

	inputTransformInputType.AddEventListener("change", func(dom.Object) {
		focusedTransform.InputType = inputTransformInputType.Get("value").String()
	})
	inputTransformOutputType.AddEventListener("change", func(dom.Object) {
		focusedTransform.OutputType = inputTransformOutputType.Get("value").String()
	})
	linkTransformFormat.AddEventListener("click", formatHandler(transformBodySession))
}

func transformImportsChange(dom.Object) {
	focusedTransform.Imports = stripCR(strings.Split(transformImportsSession.Value(), "\n"))
}

func transformBodyChange(dom.Object) {
	focusedTransform.Body = stripCR(strings.Split(transformBodySession.Value(), "\n"))
}

func (t *Transform) GainFocus() {
	focusedTransform = t
	inputTransformInputType.Set("value", t.InputType)
	inputTransformOutputType.Set("value", t.OutputType)
	transformImportsSession.SetValue(strings.Join(t.Imports, "\n"))
	transformBodySession.SetValue(strings.Join(t.Body, "\n"))
}
