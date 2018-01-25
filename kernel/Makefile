.PHONY: build run clean

# download marshmallow-x86-r3 source code
# mkdir ~/marshmallow-x86-r3; cd ~/marshmallow-x86-r3
# repo init -u git://git.osdn.net/gitroot/android-x86/manifest -b android-x86-6.0-r3
# repo sync --no-tags --no-clone-bundle
# SOURCE=`pwd`/kernel-source
# 
IMAGE=zjsyhjh/marshmallow-x86-kernel-build-env
NAME=kernel-build-container
MEM=16g
CPUS=0-19

default: run

build: Dockerfile
	docker build -f Dockerfile -t $(IMAGE) .

run: build
	docker run -it --name $(NAME) --memory=$(MEM) --cpuset-cpus=$(CPUS) $(IMAGE)

clean:
	docker stop $(NAME)
	docker rm $(NAME)
