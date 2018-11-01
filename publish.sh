#!/bin/sh -v
./test.sh
./package.sh

mv -v package/sse.zip ~/Dropbox/share/sse/

echo "Make sure dropbox is running!"
