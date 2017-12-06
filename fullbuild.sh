#!/bin/bash
set -e
set -u
cd $GOPATH/src/github.com/google/shenzhen-go

# UI JS generation
pushd view/svg
go generate
popd

# gRPC generation
protoc -I./proto shenzhen-go.proto --go_out=plugins=grpc:./api --gopherjs_out=plugins=grpc:./api
# Since both implementations live in the same package, use tags to separate them.
echo -e "//+build js\n$(cat api/shenzhen-go.pb.gopherjs.go)" > api/shenzhen-go.pb.gopherjs.go
echo -e "//+build !js\n$(cat api/shenzhen-go.pb.go)" > api/shenzhen-go.pb.go

go install github.com/google/shenzhen-go/cmd/shenzhen-go
