#!/bin/bash

if [ -z "$GOPATH" ]; then
	export GOPATH="$HOME/go"
fi

pushd $GOPATH/src/github.com/google/shenzhen-go/dev

# Client JS generation & embedding
pushd ./client
go generate
popd

# Static file embedding
pushd ./server/view
go generate
popd

go install github.com/google/shenzhen-go/dev/cmd/shenzhen-go

popd