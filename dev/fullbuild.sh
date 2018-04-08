#!/bin/bash

if [ -z "$GOPATH" ]; then
	export GOPATH="$HOME/go"
fi

# gRPC generation
pushd $GOPATH/src/github.com/google/shenzhen-go/dev/proto

go generate

cd ..

./clientbuild.sh

popd
