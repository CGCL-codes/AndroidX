# run after building marshmallow-x86, see build-marshmallow-x86/Makefile
FROM scratch
WORKDIR /

ENV OUT=~/marshmallow-x86-r3/out/target/product/x86_64
ENV KERNEL_SOURCE=marshmallow-x86-r3
COPY ${OUT}/kernel /kernel
RUN printf "KERNEL_SOURCE=${KERNEL_SOURCE}\n" > /kernel-source-info

ENTRYPOINT []
CMD []



