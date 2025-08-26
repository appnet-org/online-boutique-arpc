#!/bin/bash

. ./config.sh

set -ex

EXEC=docker

USER="appnetorg"

TAG="latest"

# for i in productpage ratings reviews details
for i in online-boutique-arpc
do
  IMAGE=${i}
  echo Processing image ${IMAGE}
  $EXEC build -t "$USER"/"$IMAGE":"$TAG" -f Dockerfile .
  $EXEC push "$USER"/"$IMAGE":"$TAG"
  echo
done