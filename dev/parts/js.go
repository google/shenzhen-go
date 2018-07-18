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
	"go/format"
	"log"

	"github.com/google/shenzhen-go/dev/dom"
)

var (
	doc = dom.CurrentDocument()
	ace = dom.GlobalAce()

	aceTheme = dom.Global("aceTheme").String()
)

func setupAce(id, mode string, handler func(dom.Object)) *dom.AceSession {
	e := ace.Edit(id)
	if e == nil {
		log.Fatalf("Couldn't ace.edit(%q)", id)
	}
	e.SetTheme("ace/theme/" + aceTheme)
	return e.Session().
		SetMode(mode).
		SetUseSoftTabs(false).
		On("change", handler)
}

func formatHandler(session *dom.AceSession) func(dom.Object) {
	return func(dom.Object) {
		buf, err := format.Source([]byte(session.Value()))
		if err != nil {
			log.Printf("Couldn't format: %v", err)
			return
		}
		session.SetValue(string(buf))
	}
}
