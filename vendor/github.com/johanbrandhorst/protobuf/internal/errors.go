package internal

import "google.golang.org/grpc/codes"

// As per https://developer.mozilla.org/en-US/docs/Web/API/CloseEvent
// codes between 4000-4999 are available for use by applications.
// Take gRPC Codes and shift by 4000 to transmit the gRPC Error over websockets.

// FormatErrorCode takes a gRPC Code and "encodes" it
// for use over the websocket bridge
func FormatErrorCode(c codes.Code) int {
	return 4000 + int(c)
}

// IsgRPCErrorCode says if the input websocket code is
// a gRPC Error code.
func IsgRPCErrorCode(i int) bool {
	return i >= 4000
}

// ParseErrorCode takes a websocket error code
// and "parses" it into a gRPC Code
func ParseErrorCode(i int) codes.Code {
	return codes.Code(i - 4000)
}
