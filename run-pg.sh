#!/bin/sh -
go run sdv.go --display-name pg-test --driver pg --db "postgres://ssetestusr:ssetestusr@localhost/ssetest" --listen-on-port 8086 --live &
wait
