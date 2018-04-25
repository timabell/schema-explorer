#!/bin/sh
psql -d sse-test < test-db.sql
echo 'test pg db created'
