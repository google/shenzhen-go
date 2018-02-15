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

// Client encapsulates all gRPC calls to a
// host-service combination.
type Client struct {
	host    string
	service string
}

// NewClient creates a new Client.
func NewClient(host, service string, opts ...DialOption) *Client {
	c := &Client{
		host:    host,
		service: service,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// RPCCall performs a unary call to an endpoint, blocking until a
// reply has been received or the context was canceled.
func (c Client) RPCCall(ctx context.Context, method string, req []byte, opts ...CallOption) ([]byte, error) {
	respChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	onMsg := func(in []byte) {
		respChan <- in
	}
	onEnd := func(s *status.Status) {
		if s.Code != codes.OK {
			errChan <- s
		} else {
			errChan <- io.EOF // Success!
		}
	}
	cancel, err := invoke(ctx, c.host, c.service, method, req, onMsg, onEnd, opts...)
	if err != nil {
		return nil, err
	}
	defer cancel()

	select {
	case err := <-errChan:
		// Wait until we've gotten the result from onEnd
		if err == io.EOF {
			select {
			// Now check for the response - should already be
			// here, but can't be too careful
			case resp := <-respChan:
				return resp, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
