#!/bin/sh
go build -o bin/linux/sse-linux-x64 sse.go
cp -r templates bin/linux/
cp -r static bin/linux/
cp peek-config.txt bin/linux/
