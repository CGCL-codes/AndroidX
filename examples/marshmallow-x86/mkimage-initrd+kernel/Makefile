.PHONY: hash tag run clean


BASE=alpine:3.6
IMAGE=ihge2k/mk-initrd
SRC=~/new-initrd
DST=/new-initrd

default: tag

hash: Dockerfile mk-initrd
	DOCKER_CONTENT_TRUST=1 docker pull $(BASE)
	docker build -t $(IMAGE):build .
	docker run --rm --entrypoint /bin/sh $(IMAGE):build -c 'cat /lib/apk/db/installed | sha1sum' | sed 's/ .*//' > $@
	

push: hash
	docker pull $(IMAGE):$(shell cat hash) || \
		(docker tag $(IMAGE):build $(IMAGE):$(shell cat hash) && \
		docker push $(IMAGE):$(shell cat hash))
	docker rmi $(IMAGE):build


tag: hash
	docker pull $(IMAGE):$(shell cat hash) || \
		docker tag $(IMAGE):build $(IMAGE):$(shell cat hash)
	docker rmi $(IMAGE):build


# docker run -it -v ~/new-initrd:/new-initrd $(IMAGE):$(shell cat hash)
run: hash
	docker pull $(IMAGE):$(shell cat hash) || \
		docker tag $(IMAGE):build $(IMAGE):$(shell cat hash)
	docker rmi $(IMAGE):build
	rm -rf $(SRC)
	mkdir $(SRC)
	docker run -it -v $(SRC):$(DST) $(IMAGE):$(shell cat hash)

clean:
	rm -f hash



