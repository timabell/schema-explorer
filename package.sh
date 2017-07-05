#!/bin/sh -v
./clean.sh
mkdir -p package/sdv/
./build.sh
./build-win.sh
cp -r README.md scripts/* bin/* package/sdv/

cd package
zip -rq sdv.zip sdv
