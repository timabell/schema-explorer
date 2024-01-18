#!/bin/sh
set -e
go test -tags "skip_mysql skip_sqlite skip_mssql" sse_test.go
