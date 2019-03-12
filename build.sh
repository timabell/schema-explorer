#!/bin/sh
go build -ldflags "-X bitbucket.org/timabell/sql-data-viewer/about.gitVersion=`git rev-parse HEAD`" -o bin/linux/schemaexplorer sse.go
cp -r templates static config bin/linux/
