#!/bin/sh
# doesn't work, we need the C dependencies. Gave up and used an ubuntu docker base instead of scratch.
CGO_ENABLED=0 GOOS=linux go build -o bin/docker/sdv-linux-cgo-x64 sdv.go
