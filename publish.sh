#!/bin/sh -v
./package.sh

# for manual downloads from https://blog.timwise.co.uk/sdv/sdv-download/
mv -v package/sdv.zip ~/Dropbox/share/sdv/

echo "Make sure dropbox is running!"
