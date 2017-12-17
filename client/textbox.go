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

// +build js

package main

import (
	"fmt"

	"github.com/google/shenzhen-go/jsutil"
)

const (
	textBoxMargin = 20
	textBoxOffset = 5
)

type textBox struct {
	d                           *diagram
	group, rect, text, textNode *jsutil.Element
	width, minWidth             float64
}

func newTextBox(d *diagram, text, textStyle, rectStyle string, x, y, minWidth, height float64) *textBox {
	b := &textBox{
		d:        d,
		group:    jsutil.MakeSVGElement("g"),
		rect:     jsutil.MakeSVGElement("rect"),
		text:     jsutil.MakeSVGElement("text"),
		textNode: &jsutil.Element{Object: jsutil.MustGetGlobal("document").Call("createTextNode", text)},
		minWidth: minWidth,
	}

	d.AddChildren(
		b.group.
			SetAttribute("transform", fmt.Sprintf("translate(%f, %f)", x, y)).
			AddChildren(
				b.rect.
					SetAttribute("height", height).
					SetAttribute("style", rectStyle),
				b.text.
					SetAttribute("y", height/2+nodeTextOffset).
					SetAttribute("text-anchor", "middle").
					SetAttribute("unselectable", "on").
					SetAttribute("style", textStyle).
					AddChildren(
						b.textNode,
					),
			),
	)
	b.computeWidth()
	return b
}

func (b *textBox) show() {
	b.group.Show()
}

func (b *textBox) hide() {
	b.group.Hide()
}

func (b *textBox) moveTo(x, y float64) {
	tf := b.group.Get("transform").Get("baseVal").Call("getItem", 0).Get("matrix")
	tf.Set("e", x)
	tf.Set("f", y)
}

func (b *textBox) setText(text string) {
	b.textNode.Set("nodeValue", text)
	b.computeWidth()
}

func (b *textBox) computeWidth() {
	b.width = b.minWidth
	if w := b.text.Call("getComputedTextLength").Float() + 2*textBoxMargin; b.minWidth < w {
		b.width = w
	}
	b.rect.SetAttribute("width", b.width)
	b.text.SetAttribute("x", b.width/2)
}
