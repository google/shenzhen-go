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

// Object is some stuff JS objects can do. This is essentially an extracted interface of *js.Object.
type Object interface {
	Get(string) Object
	Set(string, interface{})
	Delete(string)
	Length() int
	Index(int) Object
	SetIndex(int, interface{})
	Call(string, ...interface{}) Object
	Invoke(...interface{}) Object
	New(...interface{}) Object
	Bool() bool
	String() string
	Int() int
	Int64() int64
	Uint64() uint64
	Float() float64
	Interface() interface{}
	Unsafe() uintptr
}

type object struct{ *js.Object }

// WrapObject returns a wrapper for *js.Object that conforms to Object.
func WrapObject(o *js.Object) Object { return object{Object: o} }

func (o object) Get(prop string) Object              { return WrapObject(o.Object.Get(prop)) }
func (o object) Index(i int) Object                  { return WrapObject(o.Object.Index(i)) }
func (o object) Invoke(params ...interface{}) Object { return WrapObject(o.Object.Invoke(params...)) }
func (o object) New(params ...interface{}) Object    { return WrapObject(o.Object.New(params...)) }
func (o object) Call(method string, params ...interface{}) Object {
	return WrapObject(o.Object.Call(method, params...))
}
