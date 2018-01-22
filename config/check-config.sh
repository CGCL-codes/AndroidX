#!/bin/bash

#./check-config.sh android-base.cfg my-android_x86_64.cfg
if [ ! -f "$2" ]; then
    echo "$2 does not exists"
    exit 0
fi

for line in `cat $1`
do
    grep -q ${line} $2 && echo "${line} in $2" || echo "${line}" >> miss.cfg
done
