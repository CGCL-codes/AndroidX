FROM alpine:3.6

RUN \
    apk update && apk upgrade && \
    apk add --no-cache \
    cpio \
    && true


COPY ./images /images

COPY ./mk-initrd /mk-initrd

ENTRYPOINT [ "/mk-initrd" ]

