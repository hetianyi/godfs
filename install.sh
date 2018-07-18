#!/bin/sh
GOPATH=$PWD
export GOPATH

if [ ! -e "./bin" ];then
    mkdir ./bin
fi

go build -i -o $PWD/bin/storage ./src/storage.go