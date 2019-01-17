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

echo step 1/6: create build output directory.
IF NOT EXIST bin mkdir bin

echo step 2/6: install libs...
go get github.com/mattn/go-sqlite3
go get github.com/jinzhu/gorm
go get github.com/json-iterator/go
go get github.com/urfave/cli

set GOPATH=%gp%;%pwd%

echo step 3/6: build tracker...
go build -i -o bin/tracker.exe src/tracker.go

echo step 4/6: build storage...
go build -i -o bin/storage.exe src/storage.go

REM echo step 5/6: build dashboard...
REM build -i -o bin/dashboard.exe src/dashboard.go

echo step 6/6: build client...
go build -i -o bin/client.exe src/client.go
echo build success!
