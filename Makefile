REGISTRY?=chankh
IMAGE?=k8s-cloudwatch-adapter
TEMP_DIR:=$(shell mktemp -d /tmp/$(IMAGE).XXXXXX)
ARCH?=amd64
PLATFORMS=linux/arm64/v8,linux/amd64
OUT_DIR?=./_output
VENDOR_DOCKERIZED?=0
GIT_HASH?=$(shell git rev-parse --short HEAD)
TAG:=$(or ${TRAVIS_TAG},${TRAVIS_TAG},latest)
GOIMAGE=golang:1.14
GOFLAGS=-mod=vendor -tags=netgo
SRC:=$(shell find pkg cmd -type f -name "*.go")

.PHONY: all docker-build build docker docker-multiarch push test

all: verify-apis test $(OUT_DIR)/$(ARCH)/adapter

$(OUT_DIR)/%/adapter: $(SRC)
	CGO_ENABLED=0 GOARCH=$* go build $(GOFLAGS) -o $(OUT_DIR)/$*/adapter cmd/adapter/adapter.go

docker-build: verify-apis test
	cp deploy/Dockerfile $(TEMP_DIR)/Dockerfile

	docker run --rm -v $(TEMP_DIR):/build -v $(shell pwd):/go/src/github.com/awslabs/k8s-cloudwatch-adapter -e GOARCH=$(ARCH) -e GOFLAGS="$(GOFLAGS)" -w /go/src/github.com/awslabs/k8s-cloudwatch-adapter $(GOIMAGE) /bin/bash -c "\
		CGO_ENABLED=0 GO111MODULE=on go build -o /build/adapter cmd/adapter/adapter.go"

	docker build -t $(REGISTRY)/$(IMAGE):$(VERSION) $(TEMP_DIR)
	rm -rf $(TEMP_DIR)

build: $(OUT_DIR)/$(ARCH)/adapter

docker: verify-apis test
	docker build --pull -t $(REGISTRY)/$(IMAGE):$(TAG) .

docker-multiarch: verify-apis test
	docker buildx build --pull --push --platform $(PLATFORMS) --tag $(REGISTRY)/$(IMAGE):$(TAG) .

push: docker
	docker push $(REGISTRY)/$(IMAGE):$(TAG)

vendor: go.mod
ifeq ($(VENDOR_DOCKERIZED),1)
	docker run -it -v $(shell pwd):/src/k8s-cloudwatch-adapter -w /src/k8s-cloudwatch-adapter $(GOIMAGE) /bin/bash -c "\
		go mod vendor"
else
	go mod vendor
endif

test:
	CGO_ENABLED=0 GO111MODULE=on go test -cover ./pkg/...

clean:
	rm -rf ${OUT_DIR} vendor

# Code gen helpers
gen-apis: vendor
	hack/update-codegen.sh

verify-apis: vendor
	hack/verify-codegen.sh
