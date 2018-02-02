FROM alpine:3.6

RUN \
    apk update && apk upgrade && \
    apk add --no-cache \
    cpio \
    && true

WORKDIR /root

COPY initrd /root/initrd
COPY ramdisk /root/ramdisk
COPY mk-initrd.sh /root/mk-initrd.sh
COPY init /root/init
