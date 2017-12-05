#!/bin/bash
set -u
set -x
cd $GOPATH/src/github.com/google/shenzhen-go

pushd view/svg
go generate
popd

go install github.com/google/shenzhen-go/cmd/shenzhen-go
