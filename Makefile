BINARY?=bin/opcap
IMAGE_BUILDER?=podman
IMAGE_REPO?=quay.io/opdev
GIT_COMMIT=$(shell git rev-parse HEAD)
OPCAP_VERSION?="0.0.0"

PLATFORMS=linux
ARCHITECTURES=amd64 arm64 ppc64le s390x

.PHONY: build
build:
	go build -o $(BINARY) -ldflags "-X github.com/opdev/opcap/version.commit=$(GIT_COMMIT) -X github.com/opdev/opcap/version.version=$(OPCAP_VERSION)" main.go

.PHONY: build-multi-arch
build-multi-arch: $(addprefix build-linux-,$(ARCHITECTURES))

define ARCHITECTURE_template
.PHONY: build-linux-$(1)
build-linux-$(1):
	GOOS=linux GOARCH=$(1) go build -o $(BINARY)-linux-$(1) -ldflags "-X github.com/opdev/opcap/version.commit=$(GIT_COMMIT) \
				-X github.com/opdev/opcap/version.version=$(OPCAP_VERSION)" main.go
endef

$(foreach arch,$(ARCHITECTURES),$(eval $(call ARCHITECTURE_template,$(arch))))

.PHONY: fmt
fmt: gofumpt
	${GOFUMPT} -l -w .
	git diff --exit-code

.PHONY: tidy
tidy:
	go mod tidy -compat=1.17
	git diff --exit-code

.PHONY: test
test:
	go test -v $$(go list ./...) \
	-ldflags "-X github.com/opdev/opcap/version.commit=bar -X github.com/opdev/opcap/version.version=foo"

.PHONY: cover
cover:
	go test -v \
	 -ldflags "-X github.com/opdev/opcap/version.commit=bar -X github.com/opdev/opcap/version.version=foo" \
	 $$(go list ./...) \
	 -race \
	 -cover -coverprofile=coverage.out

.PHONY: vet
vet:
	go vet ./...

.PHONY: clean
clean:
	@go clean
	@# cleans the binary created by make build
	$(shell if [ -f "$(BINARY)" ]; then rm -f $(BINARY); fi)
	@# cleans all the binaries created by make build-multi-arch
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES),\
	$(shell if [ -f "$(BINARY)-$(GOOS)-$(GOARCH)" ]; then rm -f $(BINARY)-$(GOOS)-$(GOARCH); fi)))

GOFUMPT = $(shell pwd)/bin/gofumpt
gofumpt: ## Download gofumpt locally if necessary.
	$(call go-install-tool,$(GOFUMPT),mvdan.cc/gofumpt@latest)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
}
endef
