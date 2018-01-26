#!/bin/sh


DOCKERFILE=`pwd`/Dockerfile.initrd-image
HASH=`md5sum ${DOCKERFILE} | cut -d ' ' -f1`
ORG=zjsyhjh
IMAGE=androidx-init

## build init image
docker build -f ${DOCKERFILE} -t ${ORG}/${IMAGE}:${HASH} .

## push init image to docker hub
#docker push ${ORG}/${IMAGE}:${HASH}
