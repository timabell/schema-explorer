#!/bin/bash
set -e

USER=${POSTGRES_USER:-postgres}
DB=${POSTGRES_DB:-postgres}

# export PG_URL=${PG_URL:-postgresql://$USER:$POSTGRES_PASSWORD@localhost/}
echo PG_URL=$PG_URL

cd /usr/src/app/
./setup-pg-user.sh

SETUP_DB=ssetest ./setup-pg-db.sh
SETUP_DB=ssetestusr ./setup-pg-db.sh
