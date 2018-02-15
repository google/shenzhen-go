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

package metadata

import (
	"github.com/gopherjs/gopherjs/js"
	"google.golang.org/grpc/metadata"

	// Include gRPC-web JS objects
	_ "github.com/johanbrandhorst/protobuf/grpcweb/grpcwebjs"
)

// Metadata encasulates the Improbable Metadata.
type Metadata struct {
	*js.Object
	MD metadata.MD `js:"headersMap"`
}

// New initializes and populates a new Metadata.
func New(md metadata.MD) *Metadata {
	m := &Metadata{
		Object: js.Global.Get("grpc").Get("Metadata").New(),
	}
	m.MD = md
	return m
}
