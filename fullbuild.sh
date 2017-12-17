#!/bin/bash
set -u

pushd $GOPATH/src/github.com/google/shenzhen-go

# gRPC generation
rm ./proto/shenzhen-go.pb{,.gopherjs}.go
protoc -I./proto shenzhen-go.proto --go_out=plugins=grpc:./proto --gopherjs_out=plugins=grpc:./proto
# Since both implementations live in the same package, use tags to separate them.
echo -e "//+build js\n$(cat ./proto/shenzhen-go.pb.gopherjs.go)" > ./proto/shenzhen-go.pb.gopherjs.go
echo -e "//+build \x21js\n$(cat ./proto/shenzhen-go.pb.go)" > ./proto/shenzhen-go.pb.go

# Client JS generation
pushd ./client
go generate
popd

go install github.com/google/shenzhen-go/cmd/shenzhen-go

popd