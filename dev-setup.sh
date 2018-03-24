#!/bin/bash -v

# This is more a reminder of the steps than something intended to be run automatically

# Manually Download & run https://www.jetbrains.com/go/

# Install https://github.com/moovweb/gvm

sudo apt-get install curl git mercurial make binutils bison gcc build-essential

bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
source /home/tim/.gvm/scripts/gvm

gvm install go1.4 -B
gvm use go1.4
export GOROOT_BOOTSTRAP=$GOROOT
gvm install go1.9.4
gvm use go1.9.4 --default
