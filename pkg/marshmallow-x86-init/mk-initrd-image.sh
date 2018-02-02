#!/bin/sh


DOCKERFILE=`pwd`/Dockerfile.initrd-image
ORG=zjsyhjh
IMAGE=androidx-init
# md5sum
find ./ramdisk/ -type f -print0 | xargs -0 md5sum > ./md5sum-txt
find ./initrd/ -type f -print0 | xargs -0 md5sum >> ./md5sum-txt
md5sum ./mk-initrd.sh >> ./md5sum-txt
md5sum ${DOCKERFILE} >> ./md5sum-txt
HASH=`md5sum md5sum-txt | cut -d ' ' -f1` 

## build init image
docker build -f ${DOCKERFILE} -t ${ORG}/${IMAGE}:${HASH} .

## push init image to docker hub
#docker push ${ORG}/${IMAGE}:${HASH}
