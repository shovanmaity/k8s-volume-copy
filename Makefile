QUAY_USERNAME ?= k8s-volume-copy
LATEST_TAG ?= ci
IMAGE_TAG ?= $(shell git rev-parse --short HEAD)

.PHONY: crd-gen
crd-gen:
	rm -rf config/crd
	controller-gen crd:crdVersions=v1 paths=./client/apis/demo.io/v1
	controller-gen object paths=./client/apis/demo.io/v1/types.go

##
.PHONY: rsync-client-image
rsync-client-image:
	docker build -t quay.io/$(QUAY_USERNAME)/rsync-client:$(LATEST_TAG) -f docker/client/rsync/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/rsync-client:$(IMAGE_TAG) -f docker/client/rsync/Dockerfile .

.PHONY: push-rsync-client-image
push-rsync-client-image: rsync-client-image
	docker push quay.io/$(QUAY_USERNAME)/rsync-client:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/rsync-client:$(IMAGE_TAG)

##
.PHONY: rsync-daemon-image
rsync-daemon-image:
	docker build -t quay.io/$(QUAY_USERNAME)/rsync-daemon:$(LATEST_TAG) -f docker/server/rsync/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/rsync-daemon:$(IMAGE_TAG) -f docker/server/rsync/Dockerfile .

.PHONY: push-rsync-daemon-image
push-rsync-daemon-image: rsync-daemon-image
	docker push quay.io/$(QUAY_USERNAME)/rsync-daemon:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/rsync-daemon:$(IMAGE_TAG)

##
.PHONY: rsync-populator-binary
rsync-populator-binary:
	mkdir -p bin
	rm -rf bin/rsync-populator
	CGO_ENABLED=0 go build -o bin/rsync-populator app/populator/rsync/*

.PHONY: rsync-populator-image
rsync-populator-image: rsync-populator-binary
	docker build -t quay.io/$(QUAY_USERNAME)/rsync-populator:$(LATEST_TAG) -f docker/populator/rsync/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/rsync-populator:$(IMAGE_TAG) -f docker/populator/rsync/Dockerfile .

.PHONY: push-rsync-populator-image
push-rsync-populator-image: rsync-populator-image
	docker push quay.io/$(QUAY_USERNAME)/rsync-populator:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/rsync-populator:$(IMAGE_TAG)
##
.PHONY: pv-populator-binary
pv-populator-binary:
	mkdir -p bin
	rm -rf bin/pv-populator
	CGO_ENABLED=0 go build -o bin/pv-populator app/populator/pv/*

.PHONY: pv-populator-image
pv-populator-image: pv-populator-binary
	docker build -t quay.io/$(QUAY_USERNAME)/pv-populator:$(LATEST_TAG) -f docker/populator/pv/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/pv-populator:$(IMAGE_TAG) -f docker/populator/pv/Dockerfile .

.PHONY: push-pv-populator-image
push-pv-populator-image: pv-populator-image
	docker push quay.io/$(QUAY_USERNAME)/pv-populator:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/pv-populator:$(IMAGE_TAG)
##
.PHONY: images
images: rsync-client-image rsync-daemon-image rsync-populator-image pv-populator-image

.PHONY: push-images
push-images: push-rsync-client-image push-rsync-daemon-image push-rsync-populator-image push-pv-populator-image
