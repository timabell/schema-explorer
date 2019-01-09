#!/bin/sh
usr=ssetestusr
db=ssetest
# echo "removing old test db"
dropdb $db
dropuser $usr
createuser $usr
createdb $db
psql -q -c "alter user $usr with password '$usr'";
psql -q -c "alter database $db owner to $usr";
# use psql -e to echo sql along with errors while debugging
psql -d $db -q < test-db.sql
psql -d $db -q -c "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $usr;";
psql -d $db -q -c "GRANT ALL PRIVILEGES ON SCHEMA \"identity\" TO $usr;";
psql -d $db -q -c "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA \"identity\" TO $usr;";
# echo "test postgres db $db and user $usr created"
