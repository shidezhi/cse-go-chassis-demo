#!/bin/bash
set -e
set -x

BINARY_NAME="cse-go-chassis-demo-1.3.1"

CURRENT_DIR=$(cd $(dirname $0);pwd)
ROOT_PATH=$(dirname $CURRENT_DIR)

DNAME=$(dirname $(dirname "$ROOT_PATH"))
export  GOPATH="$DNAME"

cd $ROOT_PATH
if [ -f $BINARY_NAME ]; then
    rm $BINARY_NAME
fi

 CGO_ENABLED=0 go build -a -o "$BINARY_NAME"

echo "Build success!"