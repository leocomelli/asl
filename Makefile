#!/usr/bin/make

.DEFAULT_GOAL := all
PLATFORMS := linux/amd64 darwin/amd64 windows/amd64

LD_FLAGS := -ldflags "-X main.Version=`git describe --tags` -X main.BuildDate=`date -u +%Y-%m-%d_%H:%M:%S` -X main.GitHash=`git rev-parse HEAD`"

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

.PHONY: setup
setup:
	@go install github.com/securego/gosec/v2/cmd/gosec@v2.12.0
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.47.2

GOFILES=$(shell find . -type f -name '*.go' -not -path "./.git/*")

.PHONY: lint
lint:
	golangci-lint run

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
	@make -s bin test lint sec

