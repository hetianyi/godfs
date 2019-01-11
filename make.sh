#!/bin/sh

echo -e "+-------------------------------------+"
echo -e "|              >>godfs<<              |"
echo -e "| focus project on github: \033[0m           |"
echo -e "| \033[4mhttps://github.com/hetianyi/godfs\033[0m   |"
echo -e "+-------------------------------------+"

rm -rf bin

export GOROOT=$(go env GOROOT)
if [ -z "$(go env GOPATH)" ]; then
    export GOPATH=${HOME}/go:$PWD
else
    export GOPATH=$(go env GOPATH):$PWD
fi

echo "step 1/6: create build output directory."
if [ ! -e "./bin" ];then
    mkdir ./bin
fi

echo "step 2/6: install go-sqlite3 lib..."
go get github.com/mattn/go-sqlite3
go get github.com/jinzhu/gorm
go get github.com/json-iterator/go

echo "step 3/6: build tracker..."
go build -i -o $PWD/bin/tracker ./src/tracker.go

echo "step 4/6: build storage..."
go build -i -o $PWD/bin/storage ./src/storage.go

echo "step 5/6: build dashboard..."
go build -i -o $PWD/bin/dashboard ./src/dashboard.go

echo "step 6/6: build client..."
go build -i -o $PWD/bin/client ./src/client.go
echo "build success!"