#!/bin/sh

echo running mac build
env GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/timabell/schema-explorer/about.gitVersion=`git rev-parse HEAD`" -o bin/mac/schemaexplorer
cp -r templates static config bin/mac/

# also tried gox https://github.com/mitchellh/gox
# gox -cgo -output="bin/{{.Dir}}_{{.OS}}_{{.Arch}}" -os="darwin" -arch="amd64"
# but has same issues with the sqlite C code

