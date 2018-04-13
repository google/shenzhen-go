package grpcweb

import (
	"context"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/johanbrandhorst/protobuf/grpcweb/status"
)

// clientStreamClient implements an asynchronous
// writer of messages to the stream.
type clientStreamClient struct {
	ctx               context.Context
	client            *client
	errChan           chan error
	respChan          chan []byte
	isClientStreaming bool
	isServerStreaming bool
	callInfo          *callInfo
	headers           metadata.MD
	trailers          metadata.MD
}

// ClientStream is implemented by clientStreamClient
type ClientStream interface {
	Header() metadata.MD
	Trailer() metadata.MD
	SendMsg([]byte) error
	RecvMsg() ([]byte, error)
	CloseSend() error
	Context() context.Context
}

// NewClientStream performs a client-to-server streaming RPC call, returning
// a struct which exposes a Go gRPC like streaming interface.
// It is non-blocking.
func (c Client) NewClientStream(
	ctx context.Context,
	isClientStreaming,
	isServerStreaming bool,
	method string,
	opts ...CallOption,
) (ClientStream, error) {
	methodDesc := newMethodDescriptor(
		newService(c.service),
		method,
		isClientStreaming,
		isServerStreaming,
	)
	props := newProperties(c.host, isClientStreaming)
	client, err := newClient(methodDesc, props)
	if err != nil {
		return nil, status.FromError(err)
	}

	cs := &clientStreamClient{
		ctx:               ctx,
		client:            client,
		respChan:          make(chan []byte, 10),
		errChan:           make(chan error, 1),
		isClientStreaming: isClientStreaming,
		isServerStreaming: isServerStreaming,
		callInfo:          &callInfo{},
		headers:           metadata.MD{},
		trailers:          metadata.MD{},
	}

	client.onHeaders = func(headers mdwrapper) {
		// Note: we do not assign headers to the callInfo
		// as the callInfo is used in callOptions and
		// streaming callOptions cannot access headers or trailers.
		cs.headers = headers.MD
	}
	client.onMessage = func(in []byte) {
		cs.respChan <- in
	}
	client.onEnd = func(s *status.Status) {
		// Note: we do not assign trailers to the callInfo
		// as the callInfo is used in callOptions and
		// streaming callOptions cannot access headers or trailers.
		cs.trailers = s.Trailers

		// Perform CallOptions required after call
		for _, o := range opts {
			o.after(cs.callInfo)
		}

		if s.Code != codes.OK {
			cs.errChan <- s
		} else {
			cs.errChan <- io.EOF // Success!
		}
	}

	// Perform CallOptions required before call
	for _, o := range opts {
		if err := o.before(cs.callInfo); err != nil {
			return nil, status.FromError(err)
		}
	}

	md, _ := metadata.FromOutgoingContext(ctx)
	err = client.Start(md)
	if err != nil {
		return nil, status.FromError(err)
	}

	return cs, nil
}

func (cs clientStreamClient) SendMsg(payload []byte) error {
	return cs.client.Send(payload)
}

// RecvMsg blocks until either a message or an error is received.
func (cs clientStreamClient) RecvMsg() ([]byte, error) {
	select {
	case msg := <-cs.respChan:
		if !cs.isClientStreaming || cs.isServerStreaming {
			return msg, nil
		}
		// Special handling for ClientStreaming RPC, await
		// final error message. If not io.EOF, return the error.
		select {
		case err := <-cs.errChan:
			if err == io.EOF {
				return msg, nil
			}
			return nil, err
		case <-cs.ctx.Done():
			_ = cs.client.Close()

			return nil, &status.Status{
				Code:    codes.Canceled,
				Message: cs.ctx.Err().Error(),
			}
		}
	case err := <-cs.errChan:
		return nil, err
	case <-cs.ctx.Done():
		_ = cs.client.Close()

		return nil, &status.Status{
			Code:    codes.Canceled,
			Message: cs.ctx.Err().Error(),
		}
	}
}

func (cs clientStreamClient) Context() context.Context {
	return cs.ctx
}

// CloseSend is used to indicate that the client
// wants to terminate the stream.
func (cs clientStreamClient) CloseSend() error {
	return cs.client.FinishSend()
}

func (cs clientStreamClient) Header() metadata.MD {
	return cs.headers
}

func (cs clientStreamClient) Trailer() metadata.MD {
	return cs.trailers
}
