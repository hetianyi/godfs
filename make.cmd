@echo off
echo +-------------------------------------+
echo ^|              ^>^>godfs^<^<              ^|
echo ^| focus project on github:            ^|
echo ^| https://github.com/hetianyi/godfs   ^|
echo +-------------------------------------+

set GOPATH=
set GOROOT=

rmdir /Q /S bin>nul 2>nul

for /f %%i in ('chdir') do set pwd=%%i

for /f %%k in ('go env GOPATH') do set gp=%%k
for /f %%k in ('go env GOROOT') do set gr=%%k

set GOROOT=%gr%
set GOPATH=%gp%

echo gopath is %GOPATH%

echo step 1/5: create build output directory.
IF NOT EXIST bin mkdir bin

echo step 2/5: install go-sqlite3 lib...
go get github.com/mattn/go-sqlite3
set GOPATH=%gp%;%pwd%

echo step 3/5: build tracker...
go build -i -o bin/tracker.exe src/tracker.go

echo step 4/5: build storage...
go build -i -o bin/storage.exe src/storage.go

echo step 5/5: build client...
go build -i -o bin/client.exe src/client.go
echo build success!
