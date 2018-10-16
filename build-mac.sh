#!/bin/sh

echo running mac build
env CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o bin/mac/schemaexplorer
cp -r templates bin/mac/
cp -r static bin/mac/

# also tried gox https://github.com/mitchellh/gox
# gox -cgo -output="bin/{{.Dir}}_{{.OS}}_{{.Arch}}" -os="darwin" -arch="amd64"
# but has same issues with the sqlite C code

