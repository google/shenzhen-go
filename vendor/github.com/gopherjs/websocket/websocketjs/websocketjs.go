// Copyright 2014-2015 GopherJS Team. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

/*
Package websocketjs provides low-level bindings for the browser's WebSocket API.

These bindings work with typical JavaScript idioms,
such as adding event listeners with callbacks.

	ws, err := websocketjs.New("ws://localhost/socket") // Does not block.
	if err != nil {
		// handle error
	}

	onOpen := func(ev *js.Object) {
		err := ws.Send([]byte("Hello!")) // Send a binary frame.
		// ...
		err := ws.Send("Hello!") // Send a text frame.
		// ...
	}

	ws.AddEventListener("open", false, onOpen)
	ws.AddEventListener("message", false, onMessage)
	ws.AddEventListener("close", false, onClose)
	ws.AddEventListener("error", false, onError)

	err = ws.Close()
	// ...
*/
package websocketjs

import "github.com/gopherjs/gopherjs/js"

// ReadyState represents the state that a WebSocket is in. For more information
// about the available states, see
// http://dev.w3.org/html5/websockets/#dom-websocket-readystate
type ReadyState uint16

func (rs ReadyState) String() string {
	switch rs {
	case Connecting:
		return "Connecting"
	case Open:
		return "Open"
	case Closing:
		return "Closing"
	case Closed:
		return "Closed"
	default:
		return "Unknown"
	}
}

const (
	// Connecting means that the connection has not yet been established.
	Connecting ReadyState = 0
	// Open means that the WebSocket connection is established and communication
	// is possible.
	Open ReadyState = 1
	// Closing means that the connection is going through the closing handshake,
	// or the Close() method has been invoked.
	Closing ReadyState = 2
	// Closed means that the connection has been closed or could not be opened.
	Closed ReadyState = 3
)

// New creates a new low-level WebSocket. It immediately returns the new
// WebSocket.
func New(url string) (ws *WebSocket, err error) {
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if jsErr, ok := e.(*js.Error); ok && jsErr != nil {
			ws = nil
			err = jsErr
		} else {
			panic(e)
		}
	}()

	object := js.Global.Get("WebSocket").New(url)

	ws = &WebSocket{
		Object: object,
	}

	return
}

// WebSocket is a low-level convenience wrapper around the browser's WebSocket
// object. For more information, see
// http://dev.w3.org/html5/websockets/#the-websocket-interface
type WebSocket struct {
	*js.Object

	URL string `js:"url"`

	// ready state
	ReadyState     ReadyState `js:"readyState"`
	BufferedAmount uint32     `js:"bufferedAmount"`

	// networking
	Extensions string `js:"extensions"`
	Protocol   string `js:"protocol"`

	// messaging
	BinaryType string `js:"binaryType"`
}

// AddEventListener provides the ability to bind callback
// functions to the following available events:
// open, error, close, message
func (ws *WebSocket) AddEventListener(typ string, useCapture bool, listener func(*js.Object)) {
	ws.Call("addEventListener", typ, listener, useCapture)
}

// RemoveEventListener removes a previously bound callback function
func (ws *WebSocket) RemoveEventListener(typ string, useCapture bool, listener func(*js.Object)) {
	ws.Call("removeEventListener", typ, listener, useCapture)
}

// BUG(nightexcessive): When WebSocket.Send is called on a closed WebSocket, the
// thrown error doesn't seem to be caught by recover.

// Send sends a message on the WebSocket. The data argument can be a string or a
// *js.Object fulfilling the ArrayBufferView definition.
//
// See: http://dev.w3.org/html5/websockets/#dom-websocket-send
func (ws *WebSocket) Send(data interface{}) (err error) {
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if jsErr, ok := e.(*js.Error); ok && jsErr != nil {
			err = jsErr
		} else {
			panic(e)
		}
	}()
	ws.Object.Call("send", data)
	return
}

// Close closes the underlying WebSocket.
//
// See: http://dev.w3.org/html5/websockets/#dom-websocket-close
func (ws *WebSocket) Close() (err error) {
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if jsErr, ok := e.(*js.Error); ok && jsErr != nil {
			err = jsErr
		} else {
			panic(e)
		}
	}()

	// Use close code closeNormalClosure to indicate that the purpose
	// for which the connection was established has been fulfilled.
	// See https://tools.ietf.org/html/rfc6455#section-7.4.
	ws.Object.Call("close", closeNormalClosure)
	return
}

// Close codes defined in RFC 6455, section 11.7.
const (
	// 1000 indicates a normal closure, meaning that the purpose for
	// which the connection was established has been fulfilled.
	closeNormalClosure = 1000
)
