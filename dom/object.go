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

package dom

import "syscall/js"

// Object is some stuff JS objects can do. This is essentially an extracted interface of js.Value.
type Object interface {
	Get(string) Object
	Set(string, interface{})
	Length() int
	Index(int) Object
	SetIndex(int, interface{})
	Call(string, ...interface{}) Object
	Invoke(...interface{}) Object
	New(...interface{}) Object
	Bool() bool
	String() string
	Int() int
	Float() float64
}

type object struct{ js.Value }

// WrapObject returns a wrapper for js.Value that conforms to Object.
func WrapObject(o js.Value) Object { return object{Value: o} }

func (o object) Get(prop string) Object              { return WrapObject(o.Value.Get(prop)) }
func (o object) Index(i int) Object                  { return WrapObject(o.Value.Index(i)) }
func (o object) Invoke(params ...interface{}) Object { return WrapObject(o.Value.Invoke(params...)) }
func (o object) New(params ...interface{}) Object    { return WrapObject(o.Value.New(params...)) }
func (o object) Call(method string, params ...interface{}) Object {
	return WrapObject(o.Value.Call(method, params...))
}

// Global returns a name from the global namespace.
func Global(name string) Object {
	return WrapObject(js.Global().Get(name))
}
