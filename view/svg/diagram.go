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

package main

import (
	"github.com/gopherjs/gopherjs/js"
)

type diagram struct {
	*js.Object // the SVG element

	dragItem     draggable  // nil if nothing is being dragged
	selectedItem selectable // nil if nothing is selected
	errLabel     *textBox
	graph        *Graph
}

func (d *diagram) makeSVGElement(n string) *js.Object {
	return document.Call("createElementNS", d.Get("namespaceURI"), n)
}

func (d *diagram) cursorPos(e *js.Object) (x, y float64) {
	bcr := d.Call("getBoundingClientRect")
	x = e.Get("clientX").Float() - bcr.Get("left").Float()
	y = e.Get("clientY").Float() - bcr.Get("top").Float()
	return
}

func (d *diagram) mouseDown(e *js.Object) {
	if d.selectedItem == nil {
		return
	}
	d.selectedItem.loseFocus(e)
	e.Call("stopPropagation")
}

func (d *diagram) mouseMove(e *js.Object) {
	if d.dragItem == nil {
		return
	}
	d.dragItem.drag(e)
	e.Call("stopPropagation")
}

func (d *diagram) mouseUp(e *js.Object) {
	if d.dragItem == nil {
		return
	}
	d.dragItem.drag(e)
	d.dragItem.drop(e)
	d.dragItem = nil
	e.Call("stopPropagation")
}

// selecter makes an onclick handler for a selectable.
func (d *diagram) selecter(s selectable) func(*js.Object) {
	return func(e *js.Object) {
		if d.selectedItem != nil {
			d.selectedItem.loseFocus(e)
		}
		d.selectedItem = s
		s.gainFocus(e)
		e.Call("stopPropagation")
	}
}

func (d *diagram) setError(err string, x, y float64) {
	if err == "" {
		d.clearError()
		return
	}
	d.Call("appendChild", d.errLabel.group) // Bring to front
	d.errLabel.moveTo(x+4, y-36)
	d.errLabel.setText(err)
	d.errLabel.show()
}

func (d *diagram) clearError() {
	d.errLabel.hide()
}

// Point is anything that has a position on the canvas.
type Point interface {
	Pt() (x, y float64)
}

type ephemeral struct{ x, y float64 }

func (e ephemeral) Pt() (x, y float64) { return e.x, e.y }

type draggable interface {
	drag(*js.Object)
	drop(*js.Object)
}

type selectable interface {
	gainFocus(*js.Object)
	loseFocus(*js.Object)
}
