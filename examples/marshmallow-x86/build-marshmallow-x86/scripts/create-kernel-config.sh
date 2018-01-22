#!/bin/bash

SOURCE=/root/source

make -C kernel O=${SOURCE}/out/obj/kernel ARCH=x86 menuconfig
