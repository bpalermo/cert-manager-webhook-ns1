IMAGE_NAME := "cert-manager-webhook-ns1"
IMAGE_TAG := "latest"

.PHONY: build
build:
	@docker build --progress=plain -t $(IMAGE_NAME):$(IMAGE_TAG) .
