// Copyright (c) 2017 Johan Brandhorst

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package grpcweb

import "google.golang.org/grpc/metadata"

type callInfo struct {
	headers  metadata.MD
	trailers metadata.MD
}

// CallOption is a stub for any call options that may be implemented
type CallOption interface {
	before(*callInfo) error
	after(*callInfo)
}

type beforeCall func(c *callInfo) error

func (o beforeCall) before(c *callInfo) error { return o(c) }
func (o beforeCall) after(c *callInfo)        {}

type afterCall func(c *callInfo)

func (o afterCall) before(c *callInfo) error { return nil }
func (o afterCall) after(c *callInfo)        { o(c) }

// Header returns a CallOptions that retrieves the header metadata
// for a unary RPC.
func Header(headers *metadata.MD) CallOption {
	return afterCall(func(c *callInfo) {
		*headers = c.headers
	})
}

// Trailer returns a CallOptions that retrieves the trailer metadata
// for a unary RPC.
func Trailer(trailers *metadata.MD) CallOption {
	return afterCall(func(c *callInfo) {
		*trailers = c.trailers
	})
}
