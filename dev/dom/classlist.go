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

package dom

// ClassList abstracts the DOMTokenList returned by element.classList.
type ClassList interface {
	Add(class string)
	Remove(class string)
	Toggle(class string)
	Contains(class string) bool
	Replace(oldClass, newClass string)
}

type classList struct {
	Object
}

func (c classList) Add(class string)                  { c.Object.Call("add", class) }
func (c classList) Remove(class string)               { c.Object.Call("remove", class) }
func (c classList) Toggle(class string)               { c.Object.Call("toggle", class) }
func (c classList) Contains(class string) bool        { return c.Object.Call("contains", class).Bool() }
func (c classList) Replace(oldClass, newClass string) { c.Object.Call("replace", oldClass, newClass) }
