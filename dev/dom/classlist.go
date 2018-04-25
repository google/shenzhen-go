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
	Add(classes ...string)
	Remove(classes ...string)
	Toggle(class string)
	Contains(class string) bool
	Replace(oldClass, newClass string)
}

type classList struct {
	Object
}

// Probably premature optimisation.
func (c classList) oneTwoMany(f string, cs ...string) {
	switch len(cs) {
	case 0:
		return
	case 1:
		c.Object.Call(f, cs[0])
	case 2:
		c.Object.Call(f, cs[0], cs[1])
	case 3:
		c.Object.Call(f, cs[0], cs[1], cs[2])
	default:
		args := make([]interface{}, 0, len(cs))
		for _, c := range cs {
			args = append(args, c)
		}
		c.Object.Call(f, args...)
	}
}

func (c classList) Add(classes ...string) {
	c.oneTwoMany("add", classes...)
}

func (c classList) Remove(classes ...string) {
	c.oneTwoMany("remove", classes...)
}

func (c classList) Toggle(class string)               { c.Object.Call("toggle", class) }
func (c classList) Contains(class string) bool        { return c.Object.Call("contains", class).Bool() }
func (c classList) Replace(oldClass, newClass string) { c.Object.Call("replace", oldClass, newClass) }
