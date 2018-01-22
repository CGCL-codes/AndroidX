#!/bin/bash

rm -f data.img
rm -f userdata.img

data=data.img
cnt=1000

dd if=/dev/zero of=${data} bs=1M count=${cnt}
mkfs.ext4 -L data ${data}

