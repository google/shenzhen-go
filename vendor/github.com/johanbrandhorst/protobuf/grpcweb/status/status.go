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

// Package status provides a gRPC Status struct compatible
// with the Improbable gRPC-web trailers and errors.
package status

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	// Include gRPC-web JS objects
	_ "github.com/johanbrandhorst/protobuf/grpcweb/grpcwebjs"
)

// Status is a gRPC-web Status.
type Status struct {
	Code     codes.Code
	Message  string
	Trailers metadata.MD
}

// Error returns a string representation of the status
func (s Status) Error() string {
	return "rpc error: code = " + s.Code.String() + " desc = " + s.Message
}

// FromError constructs a Status from an error.
func FromError(err error) *Status {
	if err == nil {
		return nil
	}
	s, ok := err.(*Status)
	if !ok {
		s = &Status{
			Code:    codes.Unknown,
			Message: err.Error(),
		}
	}

	return s
}
