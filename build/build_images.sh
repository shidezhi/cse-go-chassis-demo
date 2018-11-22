#!/bin/bash
set -e
set -x

CURRENT_DIR=$(cd $(dirname $0);pwd)
ROOT_PATH=$(dirname $CURRENT_DIR)

cd $ROOT_PATH

IMAGE="cse-go-chassis-demo-1.3.1"
TAG="latest"
if [[ "$1" != "" ]];then
    IMAGE=$1
fi
docker build -t $IMAGE:$TAG .
#docker save -o $IMAGE-$TAG.tar $IMAGE:$TAG
