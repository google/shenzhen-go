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

import "github.com/gopherjs/gopherjs/js"

// ResponseType pretends to be a ProtobufJS ResponseType
type responseType struct {
	*js.Object
	deserializeFunc func([]byte) []byte `js:"deserializeBinary"`
}

// NewResponseType creates a new ResponseType,
// populated with a deserialization function that
// just forwards the raw bytes back to the caller.
func newResponseType() *responseType {
	r := &responseType{
		Object: js.Global.Get("Object").New(),
	}
	// Deserialization is done elsewhere
	r.deserializeFunc = func(in []byte) []byte { return in }

	return r
}

// Service pretends to be an Improbable gRPC-web Service struct.
type service struct {
	*js.Object
	name string `js:"serviceName"`
}

// NewService creates a new Service with the given name.
func newService(name string) *service {
	r := &service{
		Object: js.Global.Get("Object").New(),
	}
	r.name = name

	return r
}

// MethodDescriptor pretends to be an Improbable gRPC-web
// MethodDescriptor.
type methodDescriptor struct {
	*js.Object
	service      *service      `js:"service"`
	method       string        `js:"methodName"`
	responseType *responseType `js:"responseType"`
}

// NewMethodDescriptor creates a new MethodDescriptor.
func newMethodDescriptor(service *service, method string, responseType *responseType) *methodDescriptor {
	r := &methodDescriptor{
		Object: js.Global.Get("Object").New(),
	}
	r.service = service
	r.method = method
	r.responseType = responseType

	return r
}
