#!/bin/sh
GOROOT="/usr/local/go"
GOPATH="$PWD:/go:/go/src";
export GOROOT
export GOPATH

if [ ! -e "./bin" ];then
    mkdir ./bin
fi

go build -i -o $PWD/bin/tracker ./src/tracker.go
go build -i -o $PWD/bin/storage ./src/storage.go
#go build -i -o $PWD/bin/client ./src/client.go