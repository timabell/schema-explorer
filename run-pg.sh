#!/bin/sh -
go run sdv.go -driver pg -db "postgres://ssetest:ssetest@localhost/sse-test" -port 8086 -live &
wait
