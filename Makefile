IMG ?= softonic/ip-blocker:0.0.1-dev
CRD_OPTIONS ?= "crd:trivialVersions=true"
BIN := ip-blocker
PKG := github.com/softonic/ip-blocker
VERSION ?= 0.0.1-dev
ARCH ?= amd64
APP ?= ip-blocker
NAMESPACE ?= ip-blocker
RELEASE_NAME ?= ip-blocker
REPOSITORY ?= softonic/ip-blocker

IMAGE := $(BIN)

BUILD_IMAGE ?= golang:1.14-buster

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

.PHONY: all
all: dev

.PHONY: build
build: 
	go mod download
	GOARCH=${ARCH} go build -ldflags "-X ${PKG}/pkg/version.Version=${VERSION}" .

.PHONY: image
image:
	docker build -t $(IMG) -f Dockerfile .
	docker tag $(IMG) $(REPOSITORY):latest

.PHONY: docker-push
docker-push:
	docker push $(IMG)
	docker push $(REPOSITORY):latest

# Run tests
.PHONY: test
test: fmt vet manifests
	go test ./... -coverprofile cover.out

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: fmt vet 
	go run ./main.go

# Run go fmt against code
.PHONY: fmt
fmt:
	go fmt ./...

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

