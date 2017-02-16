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
	snapQuadrance   = 256
)

var (
	graphCanvas = js.Global.Get("document").Call("getElementById", "graph-canvas")

	ctx, canvasRect *js.Object

	width, height int

	points = make(pointSlice, 50)

	lines          []line
	possibleLine   line
	possibleLineOn bool
)

func (p point) draw(fill, stroke string) {
	ctx.Call("beginPath")
	ctx.Call("arc", p.x, p.y, 4, 0, 2*math.Pi, false)
	if fill != "" {
		ctx.Set("fillStyle", fill)
		ctx.Call("fill")
	}
	if stroke != "" {
		ctx.Set("lineWidth", 1)
		ctx.Set("strokeStyle", stroke)
		ctx.Call("stroke")
	}
}

func resize(*js.Object) {
	possibleLineOn = false

	dpr := js.Global.Get("window").Get("devicePixelRatio").Float()
	width, height = graphCanvas.Get("clientWidth").Int(), graphCanvas.Get("clientHeight").Int()
	graphCanvas.Set("width", dpr*float64(width))
	graphCanvas.Set("height", dpr*float64(height))
	ctx = graphCanvas.Call("getContext", "2d")
	ctx.Call("scale", dpr, dpr)
	canvasRect = graphCanvas.Call("getBoundingClientRect")
	redraw()
}

func (l line) draw(fill string) {
	// Line "stroke"
	ctx.Call("beginPath")
	ctx.Call("moveTo", l.p.x, l.p.y)
	ctx.Call("lineTo", l.q.x, l.q.y)
	ctx.Set("lineWidth", 4)
	ctx.Set("strokeStyle", strokeStyle)
	ctx.Call("stroke")

	// Fat dots on ends
	l.p.draw(fill, strokeStyle)
	l.q.draw(fill, strokeStyle)

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
		p.draw(normalFillStyle, strokeStyle)
	}

	for _, l := range lines {
		l.draw(normalFillStyle)
	}
	if possibleLineOn {
		possibleLine.draw(activeFillStyle)
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

		if quad, r := points.nearest(q); quad < snapQuadrance {
			possibleLine = line{r, r}
			possibleLineOn = true
			redraw()
		}
	})

	graphCanvas.Set("onmousemove", func(event *js.Object) {
		q := canvasCoord(event)
		if !possibleLineOn {
			return
		}

		possibleLine.q = q
		if quad, r := points.nearest(q); quad < snapQuadrance {
			possibleLine.q = r
		}
		redraw()
	})

	graphCanvas.Set("onmouseup", func(event *js.Object) {
		q := canvasCoord(event)
		if !possibleLineOn {
			return
		}
		if quad, r := points.nearest(q); quad < snapQuadrance {
			possibleLine.q = r
			lines = append(lines, possibleLine)
		}
		possibleLineOn = false
		redraw()
	})
}
