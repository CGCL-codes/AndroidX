FROM ubuntu:14.04

WORKDIR /root/source 
VOLUME /root/source

ENV DEBIAN_FRONTEND noninteractive

RUN dpkg --add-architecture i386

# common x86 build requirements
RUN apt-get update && \
    apt-get install -y \
    build-essential \
    curl \
    git \
    g++-multilib \
    zlib1g:i386 \ 
    libxml2-utils \
    squashfs-tools \
    dosfstools \
    mtools \
    python \
    python-libxml2 \
    python-mako \
    bison \
    zip \
    unzip \ 
    bc \
    gperf \
    gettext \
    genisoimage \
    syslinux \
    libncurses5-dev \
    libncursesw5-dev


# marshmallow-x86 build requirements
RUN apt-get update && \
  apt-get install -y software-properties-common &&\
  add-apt-repository -y ppa:openjdk-r/ppa  && \
  apt-get update && \
  apt-get install -y openjdk-7-jdk && \
  rm -rf /var/lib/apt/lists/*

