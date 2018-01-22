FROM alpine:3.6

RUN \
    apk update && apk upgrade && \
    apk add --no-cache \
    cdrkit \
    syslinux \
    && true
    


COPY mk-marshmallow-iso /mk-marshmallow-iso
# get isolinux from ~/marshmallow-x86-r3/bootable/newinstaller/boot/isolinux
COPY ./iso-bios /iso-bios

ENTRYPOINT [ "/mk-marshmallow-iso" ]

