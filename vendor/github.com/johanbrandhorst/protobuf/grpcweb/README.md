# GopherJS bindings for Improbable's gRPC-Web implementation

[![GoDoc](https://godoc.org/github.com/johanbrandhorst/protobuf/grpcweb?status.svg)](https://godoc.org/github.com/johanbrandhorst/protobuf/grpcweb)

This package provides GopherJS bindings for [Improbable's gRPC-web implementation](https://github.com/improbable-eng/grpc-web/).

It also implements a websocket client to the
[websocket-bi-directional streaming proxy](../wsproxy).

The API is still experimental, and is not currently intended for general use outside
of via the [GopherJS protoc compiler plugin](https://github.com/johanbrandhorst/protobuf/tree/master/protoc-gen-gopherjs).
See the [`protoc-gen-gopherjs` README](https://github.com/johanbrandhorst/protobuf/tree/master/protoc-gen-gopherjs/README.md)
for more information on generating the interface.
