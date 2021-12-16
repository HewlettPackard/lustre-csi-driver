
VERSION ?= $(shell sed 1q .version)
IMAGE_TAG_BASE ?= arti.dev.cray.com/lustre-csi-driver
IMG ?= $(IMAGE_TAG_BASE):$(VERSION)

image: Dockerfile
	docker build --rm --file Dockerfile --label $(IMG) --tag $(IMG) .

kind-push:
	kind load docker-image $(IMG)