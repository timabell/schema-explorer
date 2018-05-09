#!/bin/sh -v
usr=ssetestusr
db=ssetest
dropdb $db
dropuser $usr
createuser $usr
createdb $db
psql -c "alter user $usr with password '$usr'";
psql -c "alter database $db owner to $usr";
# use psql -e to echo sql along with errors while debugging
psql -d $db -e < test-db.sql
psql -d $db -c "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $usr;";
