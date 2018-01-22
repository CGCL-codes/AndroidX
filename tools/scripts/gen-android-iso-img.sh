#!/bin/bash

# ./gen-android-iso-img ISO_OUT_PATH IMAGE_PATH
# for example: ./gen-android-iso-img ~/project/android-x86_64-6.0.iso ~/project/research-project/out/images/x86`


if [ "$#" != "2" ]; then 
    echo "Usage: ./gen-android-iso-img ISO_OUT_PATH IMAGE_PATH"
    exit 1
fi

if [ ! -d "$2" ]; then
    echo "images directory is not exists"
    exit 1
fi

genisoimage -vJURT -b isolinux/isolinux.bin -c isolinux/boot.cat \
    -no-emul-boot -boot-load-size 4 -boot-info-table -eltorito-alt-boot -e boot/grub/efi.img -no-emul-boot \
    -input-charset utf-8 -V "android-x86_64-6.0" \
    -o $1 $2 
