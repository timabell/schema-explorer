#!/bin/bash
set -e

pushd . > /dev/null
cd sqlite
# test the three ways of configuring schemaexplorer
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
./test-mysql-connectionstring.sh
popd > /dev/null

pushd . > /dev/null
cd mssql
./test-mssql.sh
./test-mssql-multi-db.sh
popd > /dev/null
