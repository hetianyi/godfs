#!/bin/sh
echo "set go env."
echo "$(go env)"|awk -F= '{export $1"="$2}'
GOPATH="$PWD:/go:/go/src";
export GOPATH

echo "create build output directory."
if [ ! -e "./bin" ];then
    mkdir ./bin
fi

echo "install go-sqlite3 lib..."
go get github.com/mattn/go-sqlite3

echo "build tracker..."
go build -i -o $PWD/bin/tracker ./src/tracker.go

echo "build storage..."
go build -i -o $PWD/bin/storage ./src/storage.go

echo "build client..."
go build -i -o $PWD/bin/client ./src/client.go