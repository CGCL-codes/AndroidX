#!/bin/sh

# create new-initrd.img under ~/new-initrd folder

set -e

DST=/new-initrd

rm -rf /old-initrd
mkdir /old-initrd 
cd /old-initrd
zcat /images/initrd.img | cpio -iud
cp -f /images/ramdisk.img ./ 
cp -f /images/init ./


find . | cpio --quiet -H newc -o | gzip -9 -n > ${DST}/initrd.img


# how to run
# docker run -it -v ~/new-initrd:/new-initrd $(IMAGE) 
