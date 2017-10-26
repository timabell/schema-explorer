#!/bin/sh -v
./package.sh

# for manual downloads from https://blog.timwise.co.uk/sdv/sdv-download/
mv -v package/sdv.zip ~/Dropbox/share/sdv/

# for docker image https://github.com/timabell/sdv-docker/blob/master/Dockerfile
cp -v bin/linux/sdv-linux-x64 ~/Dropbox/share/sdv/

echo "Make sure dropbox is running!"
