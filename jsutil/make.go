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

import "github.com/gopherjs/gopherjs/js"

// Setup calls setAttribute for each pair in attrs, and appendChild for each
// element in children. It returns the element given.
func Setup(element *js.Object, attrs map[string]interface{}, events map[string]func(*js.Object), children ...*js.Object) *js.Object {
	for k, v := range attrs {
		element.Call("setAttribute", k, v)
	}
	for k, v := range events {
		element.Call("addEventListener", k, v)
	}
	for _, c := range children {
		element.Call("appendChild", c)
	}
	return element
}
