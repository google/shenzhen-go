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

package view

import (
	"fmt"

	"github.com/google/shenzhen-go/dev/dom"
)

// TextBox is an SVG group containing a filled rectangle and text.
type TextBox struct {
	// Children Rectangle and Text, and Text has child TextNode.
	dom.Element // Group
	Rectangle   dom.Element
	Text        dom.Element
	TextNode    dom.Element

	MinWidth    float64
	Margin      float64
	TextOffsetY float64
}

// MakeElements creates the DOM elements, organises them,
// and sets default attributes. Note the default is to create hidden.
func (b *TextBox) MakeElements(doc dom.Document) *TextBox {
	b.Element = doc.MakeSVGElement("g").Hide()
	b.Rectangle = doc.MakeSVGElement("rect")
	b.Text = doc.MakeSVGElement("text")
	b.TextNode = doc.MakeTextNode("")
	b.Element.
		AddChildren(b.Rectangle, b.Text)
	b.Text.
		SetAttribute("text-anchor", "middle").
		SetAttribute("unselectable", "on").
		AddChildren(b.TextNode)
	return b
}

// MoveTo moves the textbox to have the topleft corner at x, y.
func (b *TextBox) MoveTo(x, y float64) *TextBox {
	b.SetAttribute("transform", fmt.Sprintf("translate(%f, %f)", x, y))
	return b
}

// SetHeight sets the textbox height.
func (b *TextBox) SetHeight(height float64) *TextBox {
	b.Rectangle.SetAttribute("height", height)
	b.Text.SetAttribute("y", height/2+b.TextOffsetY)
	return b
}

// SetRectangleStyle sets the style of the rectangle.
func (b *TextBox) SetRectangleStyle(style string) *TextBox {
	b.Rectangle.SetAttribute("style", style)
	return b
}

// SetText sets te text in the textbox.
func (b *TextBox) SetText(text string) *TextBox {
	b.TextNode.Set("nodeValue", text)
	return b.RecomputeWidth()
}

// SetTextStyle sets the style attribute of the text.
func (b *TextBox) SetTextStyle(style string) *TextBox {
	b.Text.SetAttribute("style", style)
	return b.RecomputeWidth()
}

// SetWidth sets the width of the textbox, unless the width is less than the MinWidth,
// in which case MinWidth is used instead.
func (b *TextBox) SetWidth(w float64) *TextBox {
	if w < b.MinWidth {
		w = b.MinWidth
	}
	b.Rectangle.SetAttribute("width", w)
	b.Text.SetAttribute("x", w/2)
	return b
}

// Width returns the current width.
func (b *TextBox) Width() float64 {
	return b.Rectangle.Get("width").Float()
}

// RecomputeWidth resizes the textbox to fit all text (plus a margin).
func (b *TextBox) RecomputeWidth() *TextBox {
	return b.SetWidth(b.Text.Call("getComputedTextLength").Float() + 2*b.Margin)
}

// Remove removes the textbox from the text box's parent element.
func (b *TextBox) Remove() {
	b.Parent().RemoveChildren(b.Element)
}
