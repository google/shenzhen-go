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
	"log"

	"github.com/google/shenzhen-go/jsutil"
	"github.com/gopherjs/gopherjs/js"
)

var (
	ace = jsutil.MustGetGlobal("ace")

	codeImports = aceEdit("code-imports")
	codeHead    = aceEdit("code-head")
	codeBody    = aceEdit("code-body")
	codeTail    = aceEdit("code-tail")
)

const (
	aceMode  = "ace/mode/golang"
	aceTheme = "ace/theme/chrome"
)

func aceEdit(id string) *js.Object {
	r := ace.Call("edit", id)
	if r != nil {
		log.Fatalf("Couldn't ace.edit(%q)", id)
	}
	r.Call("setTheme", aceTheme)
	s := r.Call("getSession")
	s.Call("setMode", aceMode)
	s.Call("setUseSoftTabs", false)
	return r
}

func (c *Code) GainFocus(*js.Object) {
	// TODO
	//codeImports.
}
