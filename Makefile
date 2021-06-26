QUAY_USERNAME ?= shovanmaity
LATEST_TAG ?= ci
IMAGE_TAG ?= $(shell git rev-parse --short HEAD)

.PHONY: crd-gen
crd-gen:
	controller-gen crd:crdVersions=v1 paths=./client/apis/demo.io/v1
	controller-gen object paths=./client/apis/demo.io/v1/types.go

##
.PHONY: rsync-image
rsync-image:
	docker build -t quay.io/$(QUAY_USERNAME)/rsync:$(LATEST_TAG) -f docker/client/rsync/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/rsync:$(IMAGE_TAG) -f docker/client/rsync/Dockerfile .

.PHONY: push-rsync-image
push-rsync-image: rsync-image
	docker push quay.io/$(QUAY_USERNAME)/rsync:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/rsync:$(IMAGE_TAG)

##
.PHONY: ssh-image
ssh-image:
	docker build -t quay.io/$(QUAY_USERNAME)/ssh:$(LATEST_TAG) -f docker/client/ssh/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/ssh:$(IMAGE_TAG) -f docker/client/ssh/Dockerfile .

.PHONY: push-ssh-image
push-ssh-image: ssh-image
	docker push quay.io/$(QUAY_USERNAME)/ssh:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/ssh:$(IMAGE_TAG)

##
.PHONY: rsyncd-image
rsyncd-image:
	docker build -t quay.io/$(QUAY_USERNAME)/rsyncd:$(LATEST_TAG) -f docker/server/rsync/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/rsyncd:$(IMAGE_TAG) -f docker/server/rsync/Dockerfile .

.PHONY: push-rsyncd-image
push-rsyncd-image: rsyncd-image
	docker push quay.io/$(QUAY_USERNAME)/rsyncd:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/rsyncd:$(IMAGE_TAG)

##
.PHONY: sshd-image
sshd-image:
	docker build -t quay.io/$(QUAY_USERNAME)/sshd:$(LATEST_TAG) -f docker/server/ssh/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/sshd:$(IMAGE_TAG) -f docker/server/ssh/Dockerfile .

.PHONY: push-sshd-image
push-sshd-image: sshd-image
	docker push quay.io/$(QUAY_USERNAME)/sshd:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/sshd:$(IMAGE_TAG)

##
.PHONY: rsyncp-binary
rsyncp-binary:
	mkdir -p bin
	rm -rf bin/rsyncp
	CGO_ENABLED=0 go build -o bin/rsyncp app/populator/rsync/*

.PHONY: rsyncp-image
rsyncp-image: rsyncp-binary
	docker build -t quay.io/$(QUAY_USERNAME)/rsyncp:$(LATEST_TAG) -f docker/populator/rsync/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/rsyncp:$(IMAGE_TAG) -f docker/populator/rsync/Dockerfile .

.PHONY: push-rsyncp-image
push-rsyncp-image: rsyncp-image
	docker push quay.io/$(QUAY_USERNAME)/rsyncp:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/rsyncp:$(IMAGE_TAG)
##
.PHONY: claimp-binary
claimp-binary:
	mkdir -p bin
	rm -rf bin/claimp
	CGO_ENABLED=0 go build -o bin/claimp app/populator/claim/*

.PHONY: claimp-image
claimp-image: claimp-binary
	docker build -t quay.io/$(QUAY_USERNAME)/claimp:$(LATEST_TAG) -f docker/populator/claim/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/claimp:$(IMAGE_TAG) -f docker/populator/claim/Dockerfile .

.PHONY: push-claimp-image
push-claimp-image: claimp-image
	docker push quay.io/$(QUAY_USERNAME)/claimp:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/claimp:$(IMAGE_TAG)
##
.PHONY: images
images: rsync-image ssh-image rsyncd-image sshd-image rsyncp-image claimp-image

.PHONY: push-images
push-images: push-rsync-image push-ssh-image push-rsyncd-image push-sshd-image push-rsyncp-image push-claimp-image
