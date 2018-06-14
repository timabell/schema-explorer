#!/bin/sh -
go run sdv.go -name pg-test -driver pg -db "postgres://ssetestusr:ssetestusr@localhost/ssetest" -port 8086 -live &
wait
