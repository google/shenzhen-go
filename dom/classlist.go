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

package dom

import "syscall/js"

// ClassList provides helpful wrappers for the DOMTokenList returned by element.classList.
type ClassList struct {
	js.Value
}

// Probably premature optimisation.
func (c ClassList) oneTwoMany(f string, cs ...string) {
	switch len(cs) {
	case 0:
		return
	case 1:
		c.Call(f, cs[0])
	case 2:
		c.Call(f, cs[0], cs[1])
	case 3:
		c.Call(f, cs[0], cs[1], cs[2])
	default:
		args := make([]interface{}, 0, len(cs))
		for _, c := range cs {
			args = append(args, c)
		}
		c.Call(f, args...)
	}
}

// Add adds classes to the classlist.
func (c ClassList) Add(classes ...string) {
	c.oneTwoMany("add", classes...)
}

// Remove removes classess from the classlist.
func (c ClassList) Remove(classes ...string) {
	c.oneTwoMany("remove", classes...)
}

// Toggle calls "toggle" on the classlist.
func (c ClassList) Toggle(class string) {
	c.Call("toggle", class)
}

// Contains checks for an existing class.
func (c ClassList) Contains(class string) bool {
	return c.Call("contains", class).Bool()
}

// Replace replaces an old class with a new class.
func (c ClassList) Replace(oldClass, newClass string) {
	c.Call("replace", oldClass, newClass)
}
