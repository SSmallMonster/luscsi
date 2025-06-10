PKG = github.com/luskits/luscsi
GIT_COMMIT ?= $(shell git rev-parse HEAD)
REGISTRY ?= ghcr.io/luskits
TARGET ?= luscsi
IMAGE_NAME ?= luscsi
IMAGE_VERSION ?= 99.9-dev

IMAGE_TAG ?= $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_VERSION)
IMAGE_TAG_LATEST = $(REGISTRY)/$(IMAGE_NAME):latest
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS ?= "-X ${PKG}/pkg/luscsi.driverVersion=${IMAGE_VERSION} -X ${PKG}/pkg/luscsi.gitCommit=${GIT_COMMIT} -X ${PKG}/pkg/luscsi.buildDate=${BUILD_DATE} -extldflags '-static'"
GO111MODULE = on
GOPATH ?= $(shell go env GOPATH)
GOBIN ?= $(GOPATH)/bin
DOCKER_CLI_EXPERIMENTAL = enabled
BUILDKIT_CONFIG_FILE ?= /etc/buildkit/buildkitd.toml
export GOPATH GOBIN GO111MODULE DOCKER_CLI_EXPERIMENTAL

# The current context of image building
# The architecture of the image
ARCH ?= amd64
# Output type of docker buildx build
OUTPUT_TYPE ?= docker

DEBUG ?= true

ALL_ARCH.linux = amd64 arm64
ALL_OS_ARCH = $(foreach arch, ${ALL_ARCH.linux}, linux-$(arch))

ifeq ($(TARGET), luscsi)
build_luscsi_source_code = luscsi
dockerfile = ./build/Dockerfile
else
build_luscsi_source_code = $()
dockerfile = ./build/$(TARGET)/Dockerfile_$(TARGET)
endif

.PHONY: build
build:
ifeq ($(DEBUG), true)
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -gcflags "-N -l" -a -ldflags ${LDFLAGS} -mod vendor -o _output/luscsi.$(ARCH) ./cmd/luscsi.go
else
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -s -w -a -ldflags ${LDFLAGS} -mod vendor -o _output/luscsi.$(ARCH) ./cmd/luscsi.go
endif

.PHONY: release
release:
	docker buildx inspect container-builder || docker buildx create --use --name=container-builder

	for arch in $(ALL_ARCH.linux); do \
    		ARCH=$${arch} $(MAKE) build; \
    done

	docker buildx build --progress plain \
		--push \
		--platform linux/arm64,linux/amd64 \
		-t $(IMAGE_TAG) \
		-f $(dockerfile) .

.PHONY: render-chart-values
render-chart-values:
	@echo "Rendering chart values..."
	bash build/render-chart-values.sh

.PHONY: image
image:
	docker buildx build --pull --output=type=$(OUTPUT_TYPE) --platform="linux/$(ARCH)" \
		-t $(IMAGE_TAG)-linux-$(ARCH) --build-arg ARCH=$(ARCH) -f $(dockerfile).$(ARCH) .

.PHONY: release-manual
release-manual:
	docker buildx rm container-builder || true

	# create buildx builder
	if [ -f "$(BUILDKIT_CONFIG_FILE)" ]; then \
		docker buildx create --use --name=container-builder --config "$(BUILDKIT_CONFIG_FILE)"; \
	else \
		docker buildx create --use --name=container-builder; \
	fi


	# enable qemu for arm64 build
	# docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
	for arch in $(ALL_ARCH.linux); do \
		ARCH=$${arch} $(MAKE) build; \
		ARCH=$${arch} $(MAKE) image; \
	done

	$(MAKE) push


.PHONY: push
push:
	docker manifest create --insecure --amend $(IMAGE_TAG) $(foreach osarch, $(ALL_OS_ARCH), $(IMAGE_TAG)-${osarch})
	docker manifest push --insecure --purge $(IMAGE_TAG)
	docker manifest inspect --insecure $(IMAGE_TAG)


.PHONY: pr-image
pr-image:
	for arch in $(ALL_ARCH.linux); do \
    		ARCH=$${arch} $(MAKE) build; \
    done

	docker buildx build --progress plain \
		--load \
		--platform linux/amd64 \
		-t $(IMAGE_TAG) \
		-f $(dockerfile) .
	docker push $(IMAGE_TAG)
