#!/usr/bin/env sh

# This is a wrapper to convert environment vars into command line args,
# primarily for the docker build

echo ./sdv-linux-x64 -listenOn "$sdvListenOn" -port "$sdvPort" -driver "$sdvDriver" -db "$sdvDb"
./sdv-linux-x64 -listenOn "$sdvListenOn" -port "$sdvPort" -driver "$sdvDriver" -db "$sdvDb"
