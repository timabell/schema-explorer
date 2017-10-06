#!/bin/sh -v
rm db/test.db
sqlite3 db/test.db < test-db.sql
