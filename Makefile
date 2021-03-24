#!/usr/bin/make

.DEFAULT_GOAL := all
PLATFORMS := linux/amd64 darwin/amd64 windows/amd64

LD_FLAGS := -ldflags "-X main.Version=`git describe --tags` -X main.BuildDate=`date -u +%Y-%m-%d_%H:%M:%S` -X main.GitHash=`git rev-parse HEAD`"

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

.PHONY: setup
setup:
	@go get golang.org/x/lint/golint@v0.0.0-20201208152925-83fdc39ff7b5
	@go get golang.org/x/tools/cmd/goimports@v0.0.0-20210104081019-d8d6ddbec6ee
	@go get github.com/securego/gosec/v2@v2.5.0
	@go mod download

GOFILES=$(shell find . -type f -name '*.go' -not -path "./.git/*")

.PHONY: fmt
fmt:
	$(eval FMT_LOG := $(shell mktemp -t gofmt.XXXXX))
	@gofmt -d -s -e $(GOFILES) > $(FMT_LOG) || true
	@[ ! -s "$(FMT_LOG)" ] || (echo "gofmt failed:" | cat - $(FMT_LOG) && false)

.PHONY: imports
imports:
	$(eval IMP_LOG := $(shell mktemp -t goimp.XXXXX))
	@$(GOPATH)/bin/goimports -d -e -l $(GOFILES) > $(IMP_LOG) || true
	@[ ! -s "$(IMP_LOG)" ] || (echo "goimports failed:" | cat - $(IMP_LOG) && false)

.PHONY: lint
lint:
	@$(GOPATH)/bin/golint -set_exit_status $(shell go list ./...)

.PHONY: verify
verify:
	@make -s fmt
	@make -s imports
	@make -s lint

.PHONY: sec
sec:
	@gosec -exclude=G401,G204,G505,G101 -quiet ./...

.PHONY: bin
bin:
	go build -o ./dist/asl

.PHONY: releases
releases: $(PLATFORMS)

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build $(LD_FLAGS) -o 'dist/asl_$(os)-$(arch)'

.PHONY: test
test:
	@go test ./... -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: all
all:
	@make -s bin test verify sec

