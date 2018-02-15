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

import (
	"github.com/gopherjs/gopherjs/js"

	"github.com/johanbrandhorst/protobuf/grpcweb/metadata"
	"github.com/johanbrandhorst/protobuf/grpcweb/status"
)

// Request pretends to be an Improbable gRPC-web Request.
type request struct {
	*js.Object
	serializeFunc func() []byte `js:"serializeBinary"`
}

// NewRequest returns a new Request, populated
// with a serializeFunc that returns the bytes provided.
func newRequest(rawBytes []byte) *request {
	r := &request{
		Object: js.Global.Get("Object").New(),
	}
	r.serializeFunc = func() []byte { return rawBytes }

	return r
}

type onHeadersFunc func(metadata.Metadata)
type onEndFunc func(*status.Status)
type rawOnEndFunc func(int, string, metadata.Metadata)
type onMessageFunc func([]byte)

// Properties pretends to be an Improbable gRPC-web Properties struct.
type properties struct {
	*js.Object
	request   *request           `js:"request"`
	headers   *metadata.Metadata `js:"metadata"`
	onHeaders onHeadersFunc      `js:"onHeaders"`
	onMessage onMessageFunc      `js:"onMessage"`
	onEnd     rawOnEndFunc       `js:"onEnd"`
	host      string             `js:"host"`
	debug     bool               `js:"debug"`
}

// NewProperties creates a new, initialized, Properties struct.
func newProperties(host string, debug bool, req *request, headers *metadata.Metadata,
	onHeaders onHeadersFunc, onMsg onMessageFunc, onEnd rawOnEndFunc) *properties {
	r := &properties{
		Object: js.Global.Get("Object").New(),
	}
	r.host = host
	r.debug = debug
	r.request = req
	r.headers = headers
	r.onHeaders = onHeaders
	r.onMessage = onMsg
	r.onEnd = onEnd

	return r
}
