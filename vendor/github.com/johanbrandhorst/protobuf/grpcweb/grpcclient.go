package grpcweb

import (
	"github.com/gopherjs/gopherjs/js"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/johanbrandhorst/protobuf/grpcweb/status"
)

// mdwrapper wraps the Improbable metadata
type mdwrapper struct {
	*js.Object
	MD metadata.MD `js:"headersMap"`
}

type onHeadersFunc func(mdwrapper)
type onEndFunc func(*status.Status)
type rawOnEndFunc func(int, string, mdwrapper)
type onMessageFunc func([]byte)

type client struct {
	*js.Object
	onEnd     onEndFunc
	onMessage onMessageFunc
	onHeaders onHeadersFunc
}

func newClient(methodDesc *methodDescriptor, props *properties) (c *client, err error) {
	// Recover any thrown JS errors
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		if e, ok := e.(*js.Error); ok {
			err = e
		} else {
			panic(e)
		}
	}()

	c = &client{
		Object: js.Global.Get("grpc").Call("client", methodDesc, props),
	}
	c.Call("onHeaders", func(headers mdwrapper) {
		c.onHeaders(headers)
	})

	// Wrap onEnd in helper to translate to our status type and call
	// any post-call callOptions.
	c.Call("onEnd", func(code int, msg string, trailers mdwrapper) {
		s := &status.Status{
			Code:     codes.Code(code),
			Message:  msg,
			Trailers: trailers.MD,
		}

		c.onEnd(s)
	})

	c.Call("onMessage", func(payload []byte) {
		c.onMessage(payload)
	})

	return c, nil
}

// Close cancel any running requests on the client.
// Must only be called once, and only after the client
// has been started.
func (c client) Close() (err error) {
	// Recover any thrown JS errors
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		if e, ok := e.(*js.Error); ok {
			err = e
		} else {
			panic(e)
		}
	}()

	c.Call("close")
	return nil
}

// Start starts the client, sending
// any metadata input as headers. It
// must only be called once.
func (c client) Start(md metadata.MD) (err error) {
	// Recover any thrown JS errors
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		if e, ok := e.(*js.Error); ok {
			err = e
		} else {
			panic(e)
		}
	}()

	c.Call("start", md)
	return nil
}

func (c client) Send(payload []byte) (err error) {
	// Recover any thrown JS errors
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		if e, ok := e.(*js.Error); ok {
			err = e
		} else {
			panic(e)
		}
	}()

	c.Call("send", newRequestType(payload))
	return nil
}

// FinishSend is used for the client to stop a streaming session.
func (c client) FinishSend() (err error) {
	// Recover any thrown JS errors
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		if e, ok := e.(*js.Error); ok {
			err = e
		} else {
			panic(e)
		}
	}()

	c.Call("finishSend")
	return nil
}
