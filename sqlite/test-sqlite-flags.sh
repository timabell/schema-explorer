#!/bin/sh
set -e

echo "=================="
echo "sqlite flags"
echo "=================="

./setup.sh

# relative path hack with pwd, otherwise not resolved.
file="`pwd`/db/test.db"
cd ..
go clean -testcache
go test sse_test.go \
--driver=sqlite \
--display-name=testing-flags \
--live=true \
--listen-on-port=9999 \
--sqlite-file="$file"
#-test.v
