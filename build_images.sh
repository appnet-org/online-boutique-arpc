#!/bin/bash

. ./config.sh

set -ex

EXEC=docker

USER="appnetorg"

TAG="latest"

IMAGE="onlineboutique-arpc"

echo Processing image ${IMAGE}
$EXEC build -t "$USER"/"$IMAGE":"$TAG" -f Dockerfile .
$EXEC push "$USER"/"$IMAGE":"$TAG"