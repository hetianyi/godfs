#!/bin/sh

go build -o /bin/godfs main.go
if [ "$?" = "0" ]; then
    echo "build success!"
else
    echo "build failed!"
fi