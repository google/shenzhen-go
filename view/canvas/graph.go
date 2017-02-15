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

package main

import (
	"math"
	"math/rand"

	"github.com/gopherjs/gopherjs/js"
)

const (
	activeFillStyle = "#09f"
	normalFillStyle = "#000"
	strokeStyle     = "#ddd"
	snapLen         = 256
)

var (
	graphCanvas = js.Global.Get("document").Call("getElementById", "graph-canvas")
	ctx         = graphCanvas.Call("getContext", "2d")
	canvasRect  *js.Object

	width  int
	height int

	points = make([]point, 50)

	lines          []line
	possibleLine   line
	possibleLineOn bool
)

type point struct{ x, y int }
type line struct{ p, q point }

func resize(*js.Object) {
	possibleLineOn = false
	width, height = graphCanvas.Get("clientWidth").Int(), graphCanvas.Get("clientHeight").Int()
	graphCanvas.Set("width", width)
	graphCanvas.Set("height", height)
	canvasRect = graphCanvas.Call("getBoundingClientRect")
}

func drawLine(l line, fill string) {
	// Line outline
	ctx.Call("beginPath")
	ctx.Call("moveTo", l.p.x, l.p.y)
	ctx.Call("lineTo", l.q.x, l.q.y)
	ctx.Set("lineWidth", 4)
	ctx.Set("strokeStyle", strokeStyle)
	ctx.Call("stroke")

	// Start dot
	ctx.Call("beginPath")
	ctx.Call("arc", l.p.x, l.p.y, 4, 0, 2*math.Pi, false)
	ctx.Set("fillStyle", fill)
	ctx.Call("fill")
	ctx.Set("lineWidth", 1)
	ctx.Set("strokeStyle", strokeStyle)
	ctx.Call("stroke")

	// End dot
	ctx.Call("beginPath")
	ctx.Call("arc", l.q.x, l.q.y, 4, 0, 2*math.Pi, false)
	ctx.Set("fillStyle", fill)
	ctx.Call("fill")
	ctx.Set("lineWidth", 1)
	ctx.Set("strokeStyle", strokeStyle)
	ctx.Call("stroke")

	// Line
	ctx.Call("beginPath")
	ctx.Call("moveTo", l.p.x, l.p.y)
	ctx.Call("lineTo", l.q.x, l.q.y)
	ctx.Set("lineWidth", 2)
	ctx.Set("strokeStyle", fill)
	ctx.Call("stroke")
}

func canvasCoord(event *js.Object) point {
	return point{
		event.Get("clientX").Int() - canvasRect.Get("left").Int(),
		event.Get("clientY").Int() - canvasRect.Get("top").Int(),
	}
}

func redraw() {
	ctx.Call("clearRect", 0, 0, width, height)
	for _, p := range points {
		// Snap points
		ctx.Call("beginPath")
		ctx.Call("arc", p.x, p.y, 4, 0, 2*math.Pi, false)
		ctx.Set("fillStyle", normalFillStyle)
		ctx.Call("fill")
		ctx.Set("lineWidth", 1)
		ctx.Set("strokeStyle", strokeStyle)
		ctx.Call("stroke")
	}

	for _, l := range lines {
		drawLine(l, normalFillStyle)
	}
	if possibleLineOn {
		drawLine(possibleLine, activeFillStyle)
	}
}

func main() {
	resize(nil)
	js.Global.Get("window").Call("addEventListener", "resize", resize)

	for i := range points {
		points[i] = point{rand.Intn(width), rand.Intn(height)}
	}
	redraw()

	graphCanvas.Set("onmousedown", func(event *js.Object) {
		q := canvasCoord(event)
		for _, p := range points {
			if dx, dy := q.x-p.x, q.y-p.y; dx*dx+dy*dy < snapLen {
				possibleLine = line{p, p}
				possibleLineOn = true
				redraw()
				break
			}
		}
	})

	graphCanvas.Set("onmousemove", func(event *js.Object) {
		q := canvasCoord(event)
		if !possibleLineOn {
			return
		}

		possibleLine.q = q
		for _, p := range points {
			if dx, dy := q.x-p.x, q.y-p.y; dx*dx+dy*dy < snapLen {
				possibleLine.q = p
				break
			}
		}
		redraw()
	})

	graphCanvas.Set("onmouseup", func(event *js.Object) {
		q := canvasCoord(event)
		if !possibleLineOn {
			return
		}
		for _, p := range points {
			if dx, dy := q.x-p.x, q.y-p.y; dx*dx+dy*dy < snapLen {
				possibleLine.q = p
				lines = append(lines, possibleLine)
				break
			}
		}
		possibleLineOn = false
		redraw()
	})
}
