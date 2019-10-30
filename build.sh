#!/bin/sh
go build -ldflags "-X github.com/timabell/schema-explorer/about.gitVersion=`git rev-parse HEAD`" -o bin/linux/schemaexplorer sse.go
cp -r templates static config bin/linux/
