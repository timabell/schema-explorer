#!/bin/sh
go build -o bin/linux/sse-linux-x64 sse.go
cp -r templates static config bin/linux/
