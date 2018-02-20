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
	"context"
	"io"

	"google.golang.org/grpc/codes"

	"github.com/johanbrandhorst/protobuf/grpcweb/status"
)

// streamClient implements an asynchronous
// reader of messages received on the stream.
type streamClient struct {
	ctx      context.Context
	cancel   context.CancelFunc
	messages chan []byte
	errors   chan error
}

// ServerStream is implemented by StreamClient
type ServerStream interface {
	RecvMsg() ([]byte, error)
	Context() context.Context
}

// NewServerStream performs a server-to-client streaming RPC call, returning
// a struct which exposes a Go gRPC like streaming interface.
// It is non-blocking.
func (c Client) NewServerStream(ctx context.Context, method string, req []byte, opts ...CallOption) (ServerStream, error) {
	srv := &streamClient{
		ctx:      ctx,
		messages: make(chan []byte, 10), // Buffer up to 10 messages
		errors:   make(chan error, 1),
	}

	onMsg := func(in []byte) { srv.messages <- in }
	onEnd := func(s *status.Status) {
		if s.Code != codes.OK {
			srv.errors <- s
		} else {
			srv.errors <- io.EOF
		}
	}
	cancel, err := invoke(ctx, c.host, c.service, method, req, onMsg, onEnd, opts...)
	if err != nil {
		return nil, err
	}

	srv.cancel = cancel

	return srv, nil
}

// RecvMsg blocks until either a message or an error is received.
func (s streamClient) RecvMsg() ([]byte, error) {
	select {
	case msg := <-s.messages:
		return msg, nil
	case err := <-s.errors:
		return nil, err
	case <-s.ctx.Done():
		s.cancel()
		return nil, s.ctx.Err()
	}
}

// Context returns the stream context.
func (s streamClient) Context() context.Context {
	return s.ctx
}
