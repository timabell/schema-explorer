#!/bin/sh
go build -o bin/linux/schemaexplorer sse.go
cp -r templates static config bin/linux/
