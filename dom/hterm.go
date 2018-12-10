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

// Hterm represents a global hterm object.
type Hterm struct {
	js.Value
}

// GlobalHterm gets the global hterm object.
func GlobalHterm() Hterm {
	return Hterm{js.Global().Get("hterm")}
}

// Terminal represents a Hterm Terminal.
type Terminal struct {
	js.Value
}

// NewTerminal creates a new Terminal with the given profile.
func (h Hterm) NewTerminal(profile string) Terminal {
	return Terminal{h.Get("Terminal").New(profile)}
}

// OnTerminalReady registers a callback for when the terminal is ready.
func (t Terminal) OnTerminalReady(cb func()) {
	t.Set("onTerminalReady", cb)
}

// SetAutoCR sets the auto-carriage return feature.
func (t Terminal) SetAutoCR(enable bool) {
	t.Call("setAutoCarriageReturn", enable)
}

// Decorate decorates a sacrificial DOM element with the terminal.
func (t Terminal) Decorate(e Element) {
	t.Call("decorate", e)
}

// InstallKeyboard installs the hterm keyboard handler.
func (t Terminal) InstallKeyboard() {
	t.Call("installKeyboard")
}

// ClearHome clears the terminal and returns the cursor to 0,0.
func (t Terminal) ClearHome() {
	t.Call("clearHome")
}

// IO represents a hterm IO object.
type IO struct {
	js.Value
}

// IO returns a terminal's IO.
func (t Terminal) IO() IO {
	return IO{t.Get("io")}
}

// Push pushes an IO, something something nested IO sessions, see hterm documentation.
func (io IO) Push() IO {
	return IO{io.Call("push")}
}

// Pop unpushes this IO.
func (io IO) Pop() {
	io.Call("pop")
}

// OnVTKeystroke registers a keystroke handler.
func (io IO) OnVTKeystroke(h func(string)) {
	io.Set("onVTKeystroke", func(o js.Value) {
		h(o.String())
	})
}

// SendString registers a handler for sendString.
func (io IO) SendString(h func(string)) {
	io.Set("sendString", func(o js.Value) {
		h(o.String())
	})
}

// Print prints some text to the terminal.
func (io IO) Print(s string) {
	io.Call("print", s)
}
