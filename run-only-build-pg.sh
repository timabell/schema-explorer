#!/bin/sh
go run -tags "skip_mysql skip_sqlite skip_mssql" sse.go
