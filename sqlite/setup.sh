#!/bin/sh
if [ -d db ]; then
  # echo 'removing old test db'
  rm -rf db
fi
mkdir -p db
sqlite3 db/test.db < test-db.sql
# echo 'test sqlite db created'
