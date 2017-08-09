#!/bin/sh
docker run -p 8085:8080 --env sdvDriver=sqlite --env sdvDb=/data/Chinook_Sqlite_AutoIncrementPKs.sqlite -v /home/tim/Documents/projects/sql-data-viewer:/data timabell/sdv
