#!/bin/sh
go build -o bin/sdv-linux-x64 sdv.go

# todo - cross compile for windows
# - sdv.go:17:2: C source files not allowed when not using cgo: sqlite3-binding.c
# GOOS=windows go build -o bin/sdv-windows-x64 sdv.go
