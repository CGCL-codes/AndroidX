#!/bin/sh

DOCKERFILE=`pwd`/Dockerfile.kernel-image 
HASH=`md5sum ${DOCKERFILE}`
ORG=zjsyhjh
IMAGE=AndroidX-kernel

## build kernel image
docker build -f ${DOCKERFILE} -t ${ORG}/${IMAGE}:${HASH} .

## push kernel image to docker hub
# docker push ${ORG}/${IMAGE}:${HASH}

