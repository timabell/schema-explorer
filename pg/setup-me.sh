#!/bin/sh -v
set -e
sudo -u postgres createuser $USER
sudo -u postgres createdb $USER
sudo -u postgres psql -c "alter user $USER with superuser;" # https://stackoverflow.com/a/10757486/10245
