#!/bin/sh
mkdir -p package/sdv/
cp README.md run-mssql.sh run-public.sh run-sqlite.sh bin/sdv-linux-x64 package/sdv/

tar -czvf sdv.tar.gz -C package sdv
mv -v sdv.tar.gz ~/Dropbox/share/sdv/
