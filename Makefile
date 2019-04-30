REGISTRY?=chankh
IMAGE?=k8s-cloudwatch-adapter
TEMP_DIR:=$(shell mktemp -d /tmp/$(IMAGE).XXXXXX)
OUT_DIR?=./_output
VENDOR_DOCKERIZED=0

VERSION?=latest
GOIMAGE=golang:1.11

.PHONY: all docker-build push test build-local-image

all: test $(OUT_DIR)/adapter

src_deps=$(shell find pkg cmd -type f -name "*.go")
$(OUT_DIR)/adapter: $(src_deps)
	CGO_ENABLED=0 GOARCH=$* go build -tags netgo -o $(OUT_DIR)/$*/adapter github.com/awslabs/k8s-cloudwatch-adapter/cmd/adapter

docker-build: verify-apis
	cp deploy/Dockerfile $(TEMP_DIR)/Dockerfile

	docker run -v $(TEMP_DIR):/build -v $(shell pwd):/go/src/github.com/awslabs/k8s-cloudwatch-adapter -e GOARCH=amd64 $(GOIMAGE) /bin/bash -c "\
		CGO_ENABLED=0 go build -tags netgo -o /build/adapter github.com/awslabs/k8s-cloudwatch-adapter/cmd/adapter"

	docker build -t $(REGISTRY)/$(IMAGE):$(VERSION) $(TEMP_DIR)
	rm -rf $(TEMP_DIR)

build-local-image: $(OUT_DIR)/$(ARCH)/adapter
	sed "s|BASEIMAGE|scratch|g" deploy/Dockerfile > $(TEMP_DIR)/Dockerfile
	cp  $(OUT_DIR)/$(ARCH)/adapter $(TEMP_DIR)
	cd $(TEMP_DIR)
	docker build -t $(REGISTRY)/$(IMAGE):$(VERSION) $(TEMP_DIR)
	rm -rf $(TEMP_DIR)

push: docker-build
	docker push $(REGISTRY)/$(IMAGE):$(VERSION)

vendor: Gopkg.lock
ifeq ($(VENDOR_DOCKERIZED),1)
	docker run -it -v $(shell pwd):/go/src/github.com/awslabs/k8s-cloudwatch-adapter -w /go/src/github.com/awslabs/k8s-cloudwatch-adapter $(GOIMAGE) /bin/bash -c "\
		curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh \
		&& dep ensure -vendor-only"
else
	dep ensure -vendor-only -v
endif

test:
	CGO_ENABLED=0 go test ./pkg/...

clean:
	rm -rf ${OUT_DIR}

# Code gen helpers
gen-apis: codegen-get
	hack/update-codegen.sh

verify-apis:
	hack/verify-codegen.sh

codegen-get:
	dep ensure -add k8s.io/code-generator/...
