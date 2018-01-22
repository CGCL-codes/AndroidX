#!/bin/bash

#only set for docker
export USER=$(whoami)

make -j20 iso_img TARGET_PRODUCT=android_x86_64 TARGET_KERNEL_CONFIG=my-android_x86_64.cfg

## just build kernel
#make kernel TARGET_PRODUCT=android_x86_64 TARGET_KERNEL_CONFIG=my-android_x86_64.cfg

## use a prebuilt kernel
#SOURCE=/root/source
#make -j20 iso_img TARGET_PRODUCT=android_x86 TARGET_PREBUILT_KERNEL=${SOURCE}/out/target/product/x86_64/obj/kernel
