#!/bin/sh -v
./clean.sh
mkdir -p package/sse/
./build.sh
./build-win.sh
./build-mac.sh
cp -r README.md scripts/* bin/* package/sse/
cd package
zip -rq sse.zip sse
