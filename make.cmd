@echo off
echo +-------------------------------------+
echo ^|              ^>^>godfs^<^<              ^|
echo ^| focus project on github:            ^|
echo ^| https://github.com/hetianyi/godfs   ^|
echo +-------------------------------------+


rmdir /Q /S bin
for /f %i in ('go env GOROOT') do set GOROOT=%i
for /f %i in ('go env GOPATH') do set GOPATH=%i

echo step 1/5: create build output directory.
IF NOT EXIST bin mkdir bin

echo step 2/5: install go-sqlite3 lib...
go get github.com/mattn/go-sqlite3

echo step 3/5: build tracker...
go build -i -o bin/tracker src/tracker.go

echo step 4/5: build storage...
go build -i -o bin/storage src/storage.go

echo step 5/5: build client...
go build -i -o bin/client src/client.go
echo build success!"
