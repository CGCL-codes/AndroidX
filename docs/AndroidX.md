### 1. Introduction

This document describes how to build and run AndroidX on the ordinary GNU/Linux platform.

### 2. Requiremets

- GNU/Linux platform(eg. , `Ubuntu`)
- Docker
- `golang` 、`git` and `repo`

### 3. How to build

In this section, we describe how to build the toolkit and customize your own specified images. 

#### 3.1 Build the toolkit

 If you already have `go` installed you can use the following command to build and install the AndroidKit toolkit.

```shell
go get -u github.com/CGCL-codes/AndroidX/src/cmd/androidkit
```

#### 3.2 Build images

Once you have build the tool, use the following command to build the example configuration

```shell
export PATH=$PATH:<androidkit-path>
androidkit build AndroidX.yml
```

##### Building your own customized image

To customize, copy and modify the `AndroidX.yml` to your own `file.yml`，then run `androidkit build file.yml` to generate its specified output.

##### Yaml Specification

The yaml format specifies the image to be built:

- `AndroidX-Kernel` specifies a kernel Docker image, containing a kernel and a filesystem tarball, eg containing modules.
- `AndroidX-Init` is the base `init` process Docker image, which is unpacked as the base system, containing `init`, `containerd`, `runc` and et al.

##### Image Security

AndroidX uses [notary](https://docs.docker.com/notary/getting_started/) to support Docker image security. Using this [document](https://docs.docker.com/notary/getting_started/) to install and use the notary.

### 4. How to run

Since we boot AndroidX using `AndroidX-Kernel`+`AndroidX-Init`，`AndroidX-system.img` is necessary before running.

##### Get the system image

We use Android-x86 6.0 for generating system image, use the following command to build and get the system image.

```shell
repo init -u git://git.osdn.net/gitroot/android-x86/manifest -b android-x86-6.0-r3
repo sync --no-tags --no-clone-bundle
cd <path-to-clone-dir>/example/marshmallow-x86/build-marshmallow-x86
docker build -t <repo-name>/marshmallowx-x86-build-environment .
docker run -it --name android-build-container -v <path-to-source>:/source <repo-name>/marshmallow-x86-build-environment
```

Once you have your own customized images, you can use the following command to boot AndroidX 

```shell
androidKit run
```

See `androidkit --help` for more information. In addition, we also provide some useful scripts to help you understand how AndroidX works, which can be found in `androidkit-tools/scripts` directory.





