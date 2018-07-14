#!/bin/bash

if [ -z "$GOPATH" ]; then
	export GOPATH="$HOME/go"
fi

pushd $GOPATH/src/github.com/google/shenzhen-go/dev/proto
# gRPC stubs generation
go generate

# Also build everything else
../clientbuild.sh
popd

