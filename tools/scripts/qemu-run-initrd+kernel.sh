#!/bin/bash

# ./qemu-run-initrd+kernel.sh
# kernel, initrd.img/new-initrd.img and system.sfs/system.img must be exist in current directory

if [ ! -f "kernel" ]; then
    echo "kernel file is not exist"
    exit 1
fi

if [ ! -f "initrd.img" ] && [ ! -f "new-initrd.img" ]; then
    echo "initrd.img file is not exist"
    exit 1
elif [ -f "new-initrd.img" ]; then
    initrd="new-initrd.img"
else
    initrd="initrd.img"
fi

if [ ! -f "data.img" ] && [ ! -f "userdata.img" ]; then
    echo "userdata.img file is not exist"
    exit 1
elif [ -f "userdata.img" ]; then
    data="userdata.img"
else
    data="data.img"
fi

if [ ! -f "system.sfs" ] && [ ! -f "system.img" ]; then
    echo "system.sfs file is not exist"
    exit 1
elif [ -f "system.sfs" ]; then
    system="system.sfs"
else
    system="system.img"
fi

qemu-system-x86_64 -enable-kvm \
    -smp 2 \
    -m 400 \
    -vga std \
    -kernel kernel \
    -initrd ${initrd} \
    -append "androidboot.bootchart=50" \
    -drive file=${system} \
    -drive file=${data} \
    -redir tcp:5555::5555 

