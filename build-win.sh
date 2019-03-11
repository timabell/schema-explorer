#!/bin/sh

# prereqs:
#   apt-get install gcc-mingw-w64
#   go get -v -d -t gopkg.in/natefinch/npipe.v2

# https://github.com/mattn/go-sqlite3/issues/106#issuecomment-240179249

echo running windows build, this takes a while...
env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o bin/windows/sql-data-viewer.exe
cp -r templates static config bin/windows/
