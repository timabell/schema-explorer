#!/bin/bash -v

# This is more a reminder of the steps than something intended to be run automatically

# needed to set up the test db
sudo apt install sqlite3

# ================

# Install asdf version manaager and golang plugin

# https://asdf-vm.com/
# https://github.com/kennyp/asdf-golang

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

# ================

# for windows build

# don't know why this isn't fetched with go-get
go get gopkg.in/natefinch/npipe.v2

sudo apt install gcc-mingw-w64-x86-64
