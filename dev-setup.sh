#!/bin/bash -v

# This is more a reminder of the steps than something intended to be run automatically

# needed to set up the test db
sudo apt install sqlite3

# ================

# Install https://github.com/moovweb/gvm

sudo apt-get install curl git mercurial make binutils bison gcc build-essential

bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
source /home/tim/.gvm/scripts/gvm

gvm install go1.4 -B
gvm use go1.4
export GOROOT_BOOTSTRAP=$GOROOT
gvm install go1.9.4
gvm use go1.9.4 --default

# ================

# Manually Download & run https://www.jetbrains.com/go/

# manually set goroot & gopath in goland
echo $GOROOT
# /home/tim/.gvm/gos/go1.9.4
echo $GOPATH
# /home/tim/.gvm/pkgsets/go1.9.4/global:/home/tim/repo/go

# ================

# run this from project root
go get
