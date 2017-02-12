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

//+build js

// The canvas script is for interacting with a graph in a HTML5 canvas.
// Currently all it does is let the user draw a line with mouse events.
package main

import (
	"math"

	"github.com/gopherjs/gopherjs/js"
)

const (
	lineStyle   = "#0099ff"
	strokeStyle = "#dddddd"
)

var (
	graphCanvas   = js.Global.Get("document").Call("getElementById", "graph-canvas")
	ctx           = graphCanvas.Call("getContext", "2d")
	canvasRect    = graphCanvas.Call("getBoundingClientRect")
	width, height = graphCanvas.Get("width").Int(), graphCanvas.Get("height").Int()
)

func main() {
	on := false
	startX, startY := 0, 0

	graphCanvas.Set("onmousedown", func(this *js.Object) {
		startX = this.Get("clientX").Int() - canvasRect.Get("left").Int()
		startY = this.Get("clientY").Int() - canvasRect.Get("top").Int()
		on = true
	})

	drawScene := func(this *js.Object) {
		if !on {
			return
		}
		x := this.Get("clientX").Int() - canvasRect.Get("left").Int()
		y := this.Get("clientY").Int() - canvasRect.Get("top").Int()

		ctx.Call("clearRect", 0, 0, width, height)

		// Line outline
		ctx.Call("beginPath")
		ctx.Call("moveTo", startX, startY)
		ctx.Call("lineTo", x, y)
		ctx.Set("lineWidth", 4)
		ctx.Set("strokeStyle", strokeStyle)
		ctx.Call("stroke")

		// Start dot
		ctx.Call("beginPath")
		ctx.Call("arc", startX, startY, 4, 0, 2*math.Pi, false)
		ctx.Set("fillStyle", lineStyle)
		ctx.Call("fill")
		ctx.Set("lineWidth", 1)
		ctx.Set("strokeStyle", strokeStyle)
		ctx.Call("stroke")

		// End dot
		ctx.Call("beginPath")
		ctx.Call("arc", x, y, 4, 0, 2*math.Pi, false)
		ctx.Set("fillStyle", lineStyle)
		ctx.Call("fill")
		ctx.Set("lineWidth", 1)
		ctx.Set("strokeStyle", strokeStyle)
		ctx.Call("stroke")

		// Line
		ctx.Call("beginPath")
		ctx.Call("moveTo", startX, startY)
		ctx.Call("lineTo", x, y)
		ctx.Set("lineWidth", 2)
		ctx.Set("strokeStyle", lineStyle)
		ctx.Call("stroke")
	}

	graphCanvas.Set("onmousemove", drawScene)
	graphCanvas.Set("onmouseup", func(this *js.Object) {
		drawScene(this)
		on = false
	})
}
