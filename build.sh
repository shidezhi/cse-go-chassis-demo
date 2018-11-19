#!/bin/bash
# create PATHSRC when it not exist

PWDPATH="$PWD"
PATHSRC="src/github.com/huaweicse"
if [ ! -d "$PATHSRC" ]; then
   mkdir -p  $PATHSRC
fi

# move all file and directory to src what name not src
# test use cp replace mv

cp *[!src]*  $PATHSRC
#cp conf/ Dockerfile  main.go README.md vendor/  src/

cd $PATHSRC/

cd build/
chmod +x build.sh
# run build.sh to build
./build.sh $1

mv "cse-go-chassis-demo-1.3.1" $PWDPATH