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

package jsutil

import (
	"log"

	"github.com/gopherjs/gopherjs/js"
)

// Document is the global document element.
var Document = &Element{MustGetGlobal("document")}

// MustGetGlobal wraps js.Global.Get, and exits if the element doesn't exist.
func MustGetGlobal(id string) *js.Object {
	e := js.Global.Get(id)
	if e == nil {
		log.Fatalf("Couldn't get global %q", id)
	}
	return e
}

// MustGetElement wraps document.getElementById, and exits if the element doesn't exist.
func MustGetElement(id string) *Element {
	e := Document.Call("getElementById", id)
	if e == nil {
		log.Fatalf("Couldn't get element %q", id)
	}
	return &Element{e}
}
