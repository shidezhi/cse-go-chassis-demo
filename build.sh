#!/bin/bash
# create PATHSRC when it not exist

SHELL_FOLDER=$(cd "$(dirname "$0")";pwd)

PATHSRC="src/github.com/huaweicse"
if [ ! -d "$PATHSRC" ]; then
   mkdir -p  $PATHSRC
fi

# move all file and directory to src what name not src
# test use cp replace mv

mv *[!src]*  $PATHSRC
#cp conf/ Dockerfile  main.go README.md vendor/  src/

cd $PATHSRC/build/

chmod +x build.sh
# run build.sh to build
./build.sh $1

cd ../
mv * $SHELL_FOLDER