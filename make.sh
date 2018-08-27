#!/bin/sh
echo "$(go env)"|awk -F= '{export $1"="$2}'
GOPATH="$PWD:/go:/go/src";
export GOPATH

if [ ! -e "./bin" ];then
    mkdir ./bin
fi
go get github.com/mattn/go-sqlite3
go build -i -o $PWD/bin/tracker ./src/tracker.go
go build -i -o $PWD/bin/storage ./src/storage.go
#go build -i -o $PWD/bin/client ./src/client.go