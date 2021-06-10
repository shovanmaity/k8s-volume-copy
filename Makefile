QUAY_USERNAME ?= shovanmaity
LATEST_TAG ?= latest
IMAGE_TAG ?= 0.0.1

.PHONY: client-image
client-image:
	docker build -t quay.io/$(QUAY_USERNAME)/rsync:$(LATEST_TAG) -f package/client/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/rsync:$(IMAGE_TAG) -f package/client/Dockerfile .

.PHONY: push-client-image
push-client-image: client-image
	docker push quay.io/$(QUAY_USERNAME)/rsync:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/rsync:$(IMAGE_TAG)

.PHONY: server-image
server-image:
	docker build -t quay.io/$(QUAY_USERNAME)/rsyncd:$(LATEST_TAG) -f package/server/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/rsyncd:$(IMAGE_TAG) -f package/server/Dockerfile .

.PHONY: push-server-image
push-server-image: server-image
	docker push quay.io/$(QUAY_USERNAME)/rsyncd:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/rsyncd:$(IMAGE_TAG)

.PHONY: populator-image
populator-image:
	docker build -t quay.io/$(QUAY_USERNAME)/rsyncp:$(LATEST_TAG) -f package/populator/Dockerfile .
	docker build -t quay.io/$(QUAY_USERNAME)/rsyncp:$(IMAGE_TAG) -f package/populator/Dockerfile .

.PHONY: push-populator-image
push-populator-image: populator-image
	docker push quay.io/$(QUAY_USERNAME)/rsyncp:$(LATEST_TAG)
	docker push quay.io/$(QUAY_USERNAME)/rsyncp:$(IMAGE_TAG)

.PHONY: image
image: client-image server-image populator-image

.PHONY: push-image
push-image: push-client-image push-server-image push-populator-image
