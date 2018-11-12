#!/bin/bash
# create PATHSRC when it not exist

PATHSRC="src/demo"
if [ ! -d "$PATHSRC" ]; then
   mkdir -p  $PATHSRC
fi

# move all file and directory to src what name not src
# test use cp replace mv

mv *[!src]* src/demo
#cp conf/ Dockerfile  main.go README.md vendor/  src/

cd $PATHSRC/
mv build.sh ../../

cd build/
chmod +x build.sh
# run build.sh to build
./build.sh $1

