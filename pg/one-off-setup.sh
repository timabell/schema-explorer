#!/bin/sh
sudo -u postgres createuser tim
sudo -u postgres createdb tim
sudo -u postgres psql -c 'alter user tim with superuser;' # https://stackoverflow.com/a/10757486/10245
createuser ssetest
createdb sse-test
psql -c "alter user ssetest with password 'ssetest'";
