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
	"fmt"

	"github.com/google/shenzhen-go/jsutil"
	"github.com/gopherjs/gopherjs/js"
)

const (
	textBoxMargin = 20
	textBoxOffset = 5
)

type textBox struct {
	d                           *diagram
	group, rect, text, textNode *js.Object
	width, minWidth             float64
}

func newTextBox(d *diagram, text, textStyle, rectStyle string, x, y, minWidth, height float64) *textBox {
	b := &textBox{
		d:        d,
		group:    d.makeSVGElement("g"),
		rect:     d.makeSVGElement("rect"),
		text:     d.makeSVGElement("text"),
		textNode: jsutil.MustGetGlobal("document").Call("createTextNode", text),
		minWidth: minWidth,
	}

	jsutil.Setup(d.Object, nil, nil,
		jsutil.Setup(b.group, map[string]interface{}{
			"transform": fmt.Sprintf("translate(%f, %f)", x, y),
		},
			nil,
			jsutil.Setup(b.rect, map[string]interface{}{
				"height": height,
				"style":  rectStyle,
			},
				nil),
			jsutil.Setup(b.text, map[string]interface{}{
				"y":            height/2 + nodeTextOffset,
				"text-anchor":  "middle",
				"unselectable": "on",
				"style":        textStyle,
			},
				nil),
		))
	b.computeWidth()
	return b
}

func (b *textBox) show() {
	b.group.Call("setAttribute", "display", "")
}

func (b *textBox) hide() {
	b.group.Call("setAttribute", "display", "none")
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
	b.rect.Call("setAttribute", "width", b.width)
	b.text.Call("setAttribute", "x", b.width/2)
}
