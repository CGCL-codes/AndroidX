#!/bin/sh

# create new-initrd.img under /new-initrd folder
# run this script in a container

set -e

mkdir /old-initrd
cd /old-initrd
zcat /root/initrd/initrd.img | cpio -iud

# get ramdisk.img
cd /root/ramdisk
find . | cpio --quiet -H newc -o | gzip -9 -n > /old-initrd/ramdisk.img

# update init
cp -f /root/init /old-initrd/init

# get new-initrd.img
mkdir /new-initrd
cd /old-initrd
find . | cpio --quiet -H newc -o | gzip -9 -n > /new-initrd/initrd.img

