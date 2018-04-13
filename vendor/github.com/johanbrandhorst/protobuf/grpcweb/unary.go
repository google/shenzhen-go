package grpcweb

import (
	"context"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/johanbrandhorst/protobuf/grpcweb/status"
)

// RPCCall performs a unary call to an endpoint, blocking until a
// reply has been received or the context was canceled.
func (c Client) RPCCall(ctx context.Context, method string, req []byte, opts ...CallOption) ([]byte, error) {
	respChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	ci := &callInfo{}

	methodDesc := newMethodDescriptor(
		newService(c.service),
		method,
		false,
		false,
	)
	props := newProperties(c.host, false)
	client, err := newClient(methodDesc, props)
	if err != nil {
		return nil, status.FromError(err)
	}

	client.onHeaders = func(headers mdwrapper) {
		ci.headers = headers.MD
	}
	client.onMessage = func(in []byte) {
		respChan <- in
	}
	client.onEnd = func(s *status.Status) {
		ci.trailers = s.Trailers

		// Perform CallOptions required after call
		for _, o := range opts {
			o.after(ci)
		}

		if s.Code != codes.OK {
			errChan <- s
		} else {
			errChan <- io.EOF // Success!
		}
	}

	// Perform CallOptions required before call
	for _, o := range opts {
		if err := o.before(ci); err != nil {
			return nil, status.FromError(err)
		}
	}

	md, _ := metadata.FromOutgoingContext(ctx)
	err = client.Start(md)
	if err != nil {
		return nil, status.FromError(err)
	}

	err = client.Send(req)
	if err != nil {
		return nil, status.FromError(err)
	}

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
				_ = client.Close()
				return nil, &status.Status{
					Code:    codes.Canceled,
					Message: ctx.Err().Error(),
				}
			}
		}
		return nil, err
	case <-ctx.Done():
		_ = client.Close()
		return nil, &status.Status{
			Code:    codes.Canceled,
			Message: ctx.Err().Error(),
		}
	}
}
