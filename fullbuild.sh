#!/bin/bash
set -u

pushd $GOPATH/src/github.com/google/shenzhen-go

# gRPC generation
pushd ./proto
go generate
popd

# Client JS generation & embedding
pushd ./client
go generate
popd

# Static file embedding
pushd ./server/view
go generate
popd

go install github.com/google/shenzhen-go/cmd/shenzhen-go

popd