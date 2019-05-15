#!/bin/bash
pushd . > /dev/null
cd sqlite
./test-sqlite-env.sh
./test-sqlite-flags.sh
./test-sqlite-live.sh
popd > /dev/null

pushd . > /dev/null
cd pg
./test-pg.sh
./test-pg-multi-db.sh
popd > /dev/null

pushd . > /dev/null
cd mysql
./test-mysql.sh
popd > /dev/null

pushd . > /dev/null
cd mssql
./test-mssql.sh
./test-mssql-multi-db.sh
popd > /dev/null
