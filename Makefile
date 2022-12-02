BINARY?=bin/opcap
IMAGE_BUILDER?=podman
IMAGE_REPO?=quay.io/opdev
GO_VERSION:=$(shell go version | cut -f 3 -d " ")
BUILD_TIME:=$(shell date)
GIT_USER:=$(shell git log | grep -A2 $$(git rev-list -1 HEAD) | grep Author)
GIT_COMMIT=$(shell git rev-parse HEAD)
OPCAP_VERSION?="0.2.0"

PLATFORMS=linux
ARCHITECTURES=amd64 arm64 ppc64le s390x

.PHONY: build
build:
	go build -ldflags "-X 'github.com/opdev/opcap/cmd.GitCommit=$(GIT_COMMIT)' \
		-X 'github.com/opdev/opcap/cmd.Version=$(OPCAP_VERSION)' \
		-X 'github.com/opdev/opcap/cmd.GoVersion=$(GO_VERSION)' \
		-X 'github.com/opdev/opcap/cmd.BuildTime=$(BUILD_TIME)' \
		-X 'github.com/opdev/opcap/cmd.GitUser=$(GIT_USER)'" \
		-o $(BINARY) main.go

.PHONY: build-multi-arch
build-multi-arch: $(addprefix build-linux-,$(ARCHITECTURES))

define ARCHITECTURE_template
.PHONY: build-linux-$(1)
build-linux-$(1):
	GOOS=linux GOARCH=$(1) go build -o $(BINARY)-linux-$(1) -ldflags "-X 'github.com/opdev/opcap/cmd.GitCommit=$(GIT_COMMIT)' \
				-X 'github.com/opdev/opcap/cmd.Version=$(OPCAP_VERSION)'" main.go
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
	-ldflags "-X 'github.com/opdev/opcap/cmd.GitCommit=bar' -X 'github.com/opdev/opcap/cmd.Version=foo'"

.PHONY: cover
cover:
	go test -v \
	 -ldflags "-X 'github.com/opdev/opcap/cmd.GitCommit=bar' -X 'github.com/opdev/opcap/cmd.Version=foo'" \
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
