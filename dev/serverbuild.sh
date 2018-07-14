#!/bin/bash

if [ -z "$GOPATH" ]; then
	export GOPATH="$HOME/go"
fi

# Ensure embed script exists
# TODO(josh): Do this less rudely
go install github.com/google/shenzhen-go/scripts/embed

pushd $GOPATH/src/github.com/google/shenzhen-go/dev/server/view
# Statically embed resources
go generate
popd

go install -tags webview github.com/google/shenzhen-go/dev/cmd/shenzhen-go
