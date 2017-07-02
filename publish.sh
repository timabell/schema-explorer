#!/bin/sh -v
mkdir -p package/sdv/
./build.sh
./build-win.sh
cp -r README.md scripts/* bin/* package/sdv/
tar -czvf sdv.tar.gz -C package sdv
mv -v sdv.tar.gz ~/Dropbox/share/sdv/
