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

import "github.com/google/shenzhen-go/dev/dom"

// TextBox is an SVG group containing a filled rectangle and text.
type TextBox struct {
	// Children Rectangle and Text, and Text has child TextNode.
	Group
	Rect     dom.Element
	Text     dom.Element
	TextNode dom.Element

	MinWidth float64
	Margin   float64

	currentText string
}

// MakeElements creates the DOM elements, organises them,
// and sets default attributes. The return is the main group.
func (b *TextBox) MakeElements(doc dom.Document, parent dom.Element) *TextBox {
	b.Group = NewGroup(doc, parent)
	b.Group.Element.ClassList().Add("textbox")
	b.Rect = doc.MakeSVGElement("rect")
	b.Text = doc.MakeSVGElement("text")
	b.TextNode = doc.MakeTextNode("")
	b.Group.
		AddChildren(b.Rect, b.Text)
	b.Text.
		AddChildren(b.TextNode)
	return b
}

// SetHeight sets the textbox height.
func (b *TextBox) SetHeight(height float64) *TextBox {
	b.Rect.SetAttribute("height", height)
	b.Text.SetAttribute("y", height/2)
	return b
}

// CurrentText returns the last string passed to SetText.
func (b *TextBox) CurrentText() string {
	return b.currentText
}

// SetText sets te text in the textbox.
func (b *TextBox) SetText(text string) *TextBox {
	b.TextNode.Set("nodeValue", text)
	//b.RecomputeWidth()
	b.currentText = text
	return b
}

// SetWidth sets the width of the textbox, unless the width is less than the MinWidth,
// in which case MinWidth is used instead.
func (b *TextBox) SetWidth(w float64) *TextBox {
	if w < b.MinWidth {
		w = b.MinWidth
	}
	b.Rect.SetAttribute("width", w)
	b.Text.SetAttribute("x", w/2)
	return b
}

// Width returns the current width.
func (b *TextBox) Width() float64 {
	return b.Rect.GetAttribute("width").Float()
}

// RecomputeWidth resizes the textbox to fit all text (plus a margin).
func (b *TextBox) RecomputeWidth() *TextBox {
	// This is kind of ridiculous.
	// I just want some text centred in a rect.
	// You can't add the text as a child of the rect, so we add both to a group.
	// Then to figure out how big the rect needs to be, use the size of the text.
	// getComputedTextLength does this, but only if the text has been rendered.
	// getBBox works kind of similarly, but when it works it works better, and when
	// it fails it crashes hard (on Firefox).

	//b.SetWidth(b.Text.Call("getComputedTextLength").Float() + 2*b.Margin)
	w := b.Text.Call("getBBox").Get("width").Float()
	return b.SetWidth(w + 2*b.Margin)
}
