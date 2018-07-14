#!/bin/bash

if [ -z "$GOPATH" ]; then
	export GOPATH="$HOME/go"
fi

pushd $GOPATH/src/github.com/google/shenzhen-go/dev/client
# Client JS generation
go generate

# Build everything server-related
../serverbuild.sh
popd