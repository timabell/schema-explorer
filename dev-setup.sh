#!/bin/bash -v

# This is more a reminder of the steps than something intended to be run automatically

# needed to set up the test db
sudo apt install sqlite3

# needed for pg test scripts even if pg is running in docker
sudo apt install postgresql-client
# ================

# Install asdf version manaager and golang plugin

# https://asdf-vm.com/
# https://github.com/kennyp/asdf-golang

asdf plugin add golang

# ================

# Manually Download & run https://www.jetbrains.com/go/

# manually set goroot in goland
echo $GOROOT
# e.g. /home/tim/.gvm/gos/go1.9.4

# ================

# for windows build
sudo apt install gcc-mingw-w64-x86-64

# setup docker images to run tests and try things locally

pg/docker-start.sh

(
	cd mysql
	./docker-first-run.sh
)

(
	cd mssql
	./docker-first-run.sh
)
