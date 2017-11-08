#!/bin/sh
go build -o bin/linux/sdv-linux-x64 sdv.go
cp -r templates bin/linux/
