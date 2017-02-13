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
// Currently all it does is let the user draw a line between random dots.
package main

import (
	"math"
	"math/rand"

	"github.com/gopherjs/gopherjs/js"
)

const (
	fillStyle   = "#0099ff"
	strokeStyle = "#dddddd"
	snapLen     = 225
)

var (
	graphCanvas   = js.Global.Get("document").Call("getElementById", "graph-canvas")
	ctx           = graphCanvas.Call("getContext", "2d")
	width, height int
	on            bool
)

type point struct{ x, y int }

func resize(*js.Object) {
	on = false
	width, height = graphCanvas.Get("clientWidth").Int(), graphCanvas.Get("clientHeight").Int()
	graphCanvas.Set("width", width)
	graphCanvas.Set("height", height)
}

func main() {
	resize(nil)
	js.Global.Get("window").Call("addEventListener", "resize", resize)
	canvasRect := graphCanvas.Call("getBoundingClientRect")

	startX, startY := 0, 0
	snap := make([]point, 50)
	for i := range snap {
		snap[i] = point{rand.Intn(width), rand.Intn(height)}
	}

	drawPoints := func() {
		// Snap point
		for _, p := range snap {
			ctx.Call("beginPath")
			ctx.Call("arc", p.x, p.y, 4, 0, 2*math.Pi, false)
			ctx.Set("fillStyle", "#000")
			ctx.Call("fill")
			ctx.Set("lineWidth", 1)
			ctx.Set("strokeStyle", strokeStyle)
			ctx.Call("stroke")
		}
	}
	drawPoints()

	graphCanvas.Set("onmousedown", func(event *js.Object) {
		startX = event.Get("clientX").Int() - canvasRect.Get("left").Int()
		startY = event.Get("clientY").Int() - canvasRect.Get("top").Int()
		for _, p := range snap {
			if dx, dy := startX-p.x, startY-p.y; dx*dx+dy*dy < 100 {
				startX, startY = p.x, p.y
				on = true
				break
			}
		}
	})

	drawLine := func(x1, y1, x2, y2 int) {
		// Line outline
		ctx.Call("beginPath")
		ctx.Call("moveTo", x1, y1)
		ctx.Call("lineTo", x2, y2)
		ctx.Set("lineWidth", 4)
		ctx.Set("strokeStyle", strokeStyle)
		ctx.Call("stroke")

		// Start dot
		ctx.Call("beginPath")
		ctx.Call("arc", x1, y1, 4, 0, 2*math.Pi, false)
		ctx.Set("fillStyle", fillStyle)
		ctx.Call("fill")
		ctx.Set("lineWidth", 1)
		ctx.Set("strokeStyle", strokeStyle)
		ctx.Call("stroke")

		// End dot
		ctx.Call("beginPath")
		ctx.Call("arc", x2, y2, 4, 0, 2*math.Pi, false)
		ctx.Set("fillStyle", fillStyle)
		ctx.Call("fill")
		ctx.Set("lineWidth", 1)
		ctx.Set("strokeStyle", strokeStyle)
		ctx.Call("stroke")

		// Line
		ctx.Call("beginPath")
		ctx.Call("moveTo", x1, y1)
		ctx.Call("lineTo", x2, y2)
		ctx.Set("lineWidth", 2)
		ctx.Set("strokeStyle", fillStyle)
		ctx.Call("stroke")
	}

	graphCanvas.Set("onmousemove", func(event *js.Object) {
		x := event.Get("clientX").Int() - canvasRect.Get("left").Int()
		y := event.Get("clientY").Int() - canvasRect.Get("top").Int()
		if !on {
			return
		}

		ctx.Call("clearRect", 0, 0, width, height)
		drawPoints()
		for _, p := range snap {
			if dx, dy := x-p.x, y-p.y; dx*dx+dy*dy < snapLen {
				drawLine(startX, startY, p.x, p.y)
				return
			}
		}
		drawLine(startX, startY, x, y)
	})

	graphCanvas.Set("onmouseup", func(event *js.Object) {
		x := event.Get("clientX").Int() - canvasRect.Get("left").Int()
		y := event.Get("clientY").Int() - canvasRect.Get("top").Int()
		if !on {
			return
		}
		ctx.Call("clearRect", 0, 0, width, height)
		drawPoints()
		for _, p := range snap {
			if dx, dy := x-p.x, y-p.y; -10 < dx && dx < 10 && -10 < dy && dy < 10 {
				drawLine(startX, startY, p.x, p.y)
				break
			}
		}
		on = false
	})
}
