PKG = luskits/luscsi
GIT_COMMIT ?= $(shell git rev-parse HEAD)
REGISTRY ?= ghcr.io/luskits
TARGET ?= luscsi
IMAGE_NAME ?= luscsi
IMAGE_VERSION ?= 99.9-dev

IMAGE_TAG ?= $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_VERSION)
IMAGE_TAG_LATEST = $(REGISTRY)/$(IMAGE_NAME):latest
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS ?= "-X ${PKG}/pkg/csi.driverVersion=${IMAGE_VERSION} -X ${PKG}/pkg/csi.gitCommit=${GIT_COMMIT} -X ${PKG}/pkg/csi.buildDate=${BUILD_DATE} -extldflags '-static'"
GO111MODULE = on
GOPATH ?= $(shell go env GOPATH)
GOBIN ?= $(GOPATH)/bin
DOCKER_CLI_EXPERIMENTAL = enabled
export GOPATH GOBIN GO111MODULE DOCKER_CLI_EXPERIMENTAL

# The current context of image building
# The architecture of the image
ARCH ?= amd64
# Output type of docker buildx build
OUTPUT_TYPE ?= registry

DEBUG ?= true

ALL_ARCH.linux = amd64 #arm64
ALL_OS_ARCH = $(foreach arch, ${ALL_ARCH.linux}, linux-$(arch))

ifeq ($(TARGET), luscsi)
build_luscsi_source_code = luscsi
dockerfile = ./build/Dockerfile
else
build_luscsi_source_code = $()
dockerfile = ./build/$(TARGET)/Dockerfile_$(TARGET)
endif

.PHONY: luscsi
luscsi:
ifeq ($(DEBUG), true)
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -gcflags "-N -l" -a -ldflags ${LDFLAGS} -mod vendor -o _output/luscsi ./cmd/luscsi.go
else
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -s -w -a -ldflags ${LDFLAGS} -mod vendor -o _output/luscsi ./cmd/luscsi.go
endif

.PHONY: container-linux
container-linux:
	docker buildx build --pull --output=type=$(OUTPUT_TYPE) --platform="linux/$(ARCH)" \
		-t $(IMAGE_TAG)-linux-$(ARCH) --build-arg ARCH=$(ARCH) -f $(dockerfile) .

.PHONY: luscsi-container
luscsi-container:
	docker buildx rm container-builder || true
	docker buildx create --use --name=container-builder

	# enable qemu for arm64 build
	# docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
	for arch in $(ALL_ARCH.linux); do \
		ARCH=$${arch} $(MAKE) luscsi; \
		ARCH=$${arch} $(MAKE) container-linux; \
	done

.PHONY: push
push:
ifdef CI
	docker manifest create --amend $(IMAGE_TAG) $(foreach osarch, $(ALL_OS_ARCH), $(IMAGE_TAG)-${osarch})
	docker manifest push --purge $(IMAGE_TAG)
	docker manifest inspect $(IMAGE_TAG)
else
	docker push $(IMAGE_TAG)
endif

.PHONY: build-push
build-push: $(build_luscsi_source_code)
	docker tag $(IMAGE_TAG) $(IMAGE_TAG_LATEST)
	docker push $(IMAGE_TAG_LATEST)