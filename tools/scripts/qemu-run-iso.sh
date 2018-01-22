#!/bin/bash

# ./qemu-android ISO_PATH

if [ "$#" != "1" ]; then
    echo "Usage: ./qemu-android ISO_PATH"
    exit 1
fi

if [ ! -f "$1" ]; then
    echo "iso file is not exist"
    exit 1
fi

qemu-system-x86_64 -enable-kvm \
    -smp 2 \
    -m 4096 \
    -vga std \
    -net nic \
    -net user \
    -cdrom \
    $1

