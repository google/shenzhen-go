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

package view

// Pointer is anything that has a position on the canvas.
type Pointer interface {
	Pt() (x, y float64)
}

// Point is a basic implementation of Point.
type Point struct{ x, y float64 }

// Pt implements Pointer.
func (p Point) Pt() (x, y float64) { return p.x, p.y }

// Set changes the point's value.
func (p *Point) Set(x, y float64) { p.x, p.y = x, y }

// Add performs 2D vector addition.
func (p *Point) Add(x, y float64) { p.x += x; p.y += y }

// Scale performs scalar multiplication.
func (p *Point) Scale(k float64) { p.x *= k; p.y *= k }
