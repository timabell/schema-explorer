#!/bin/sh -v
set -e # exit on error
asdf plugin update golang
latest=`asdf list all golang | grep -Ev 'rc|beta' | tail -n 1`
echo $latest
asdf install golang $latest
asdf local golang $latest
./test.sh
git commit -i .tool-versions -m "Upgrade golang to latest"
