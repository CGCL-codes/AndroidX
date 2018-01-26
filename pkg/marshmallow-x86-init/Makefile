.PHONY: build run clean

IMAGE=zjsyhjh/marshmallow-x86-initrd-build-env
NAME=initrd-build-container


default: run


build: Dockerfile
	docker build -f Dockerfile -t $(IMAGE) .

run: build
	docker run -it --name $(NAME) $(IMAGE)


clean:
	docker stop $(NAME)
	docker rm $(NAME)


