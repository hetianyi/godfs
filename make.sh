#!/bin/sh

echo -e "+-------------------------------------+"
echo -e "|              >>godfs<<              |"
echo -e "| focus project on github: \033[0m           |"
echo -e "| \033[4mhttps://github.com/hetianyi/godfs\033[0m   |"
echo -e "+-------------------------------------+"

rm -rf bin

export GOROOT=$(go env GOROOT)
export GOPATH="$(go env GOPATH):$PWD"

echo "step 1/5: create build output directory."
if [ ! -e "./bin" ];then
    mkdir ./bin
fi

echo "step 2/5: install go-sqlite3 lib..."
go get github.com/mattn/go-sqlite3

echo "step 3/5: build tracker..."
go build -i -o $PWD/bin/tracker ./src/tracker.go

echo "step 4/5: build storage..."
go build -i -o $PWD/bin/storage ./src/storage.go

echo "step 5/5: build client..."
go build -i -o $PWD/bin/client ./src/client.go
echo "build success!"