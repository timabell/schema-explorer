#!/bin/sh

echo running mac build
env CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o bin/mac/schemaexplorer
cp -r templates bin/mac/
cp -r static bin/mac/
