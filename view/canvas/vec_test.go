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
	"math/rand"
	"testing"
)

var (
	quad int
	res  point
)

func benchmarkSliceNearestN(b *testing.B, n int) {
	q, r := 0, point{}
	b.StopTimer()
	ps := make(pointSlice, n)
	for i := 0; i < n; i++ {
		ps[i] = point{rand.Int(), rand.Int()}
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q, r = ps.nearest(point{rand.Int(), rand.Int()})
	}
	quad, res = q, r
}

func BenchmarkSliceNearest10(b *testing.B)      { benchmarkSliceNearestN(b, 10) }
func BenchmarkSliceNearest100(b *testing.B)     { benchmarkSliceNearestN(b, 100) }
func BenchmarkSliceNearest1000(b *testing.B)    { benchmarkSliceNearestN(b, 1000) }
func BenchmarkSliceNearest10000(b *testing.B)   { benchmarkSliceNearestN(b, 10000) }
func BenchmarkSliceNearest100000(b *testing.B)  { benchmarkSliceNearestN(b, 100000) }
func BenchmarkSliceNearest1000000(b *testing.B) { benchmarkSliceNearestN(b, 1000000) }
