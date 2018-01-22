## Android-x86 Build on Docker Container

#### To build the container image
```shell
docker build -t zjsyhjh/marshmallow-x86-build-environment .
```

#### Marshmallow-x86 build
```shell
docker run -it --name android-build-container -v <path-to-source>:/source zjsyhjh/marshmallow-x86-build-environment
```
