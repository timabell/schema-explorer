#!/bin/sh
CGO_ENABLED=0 GOOS=linux go build -o bin/docker/sdv-linux-cgo-x64 sdv.go
