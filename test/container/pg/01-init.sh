#!/bin/bash
set -e

USER=ssetestusr
PASS=ssetestusrpass
DB=ssetest

dropdb --if-exists $DB
dropuser --if-exists $USER
createuser $USER
createdb $DB

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    ALTER USER $USER with password '$PASS';
    ALTER DATABASE $DB OWNER TO $USER;
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$DB" -q < /usr/src/app/test-db.sql

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$DB" <<-EOSQL
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $USER;
    GRANT ALL PRIVILEGES ON SCHEMA "identity" TO $USER;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA "identity" TO $USER;
EOSQL
