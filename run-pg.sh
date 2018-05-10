#!/bin/sh -
go run sdv.go -driver pg -db "postgres://ssetestusr:ssetestusr@localhost/ssetest" -port 8086 -live &
wait
