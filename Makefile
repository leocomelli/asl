#!/usr/bin/make

.DEFAULT_GOAL := all

.PHONY: setup
setup:
	@go install github.com/securego/gosec/v2/cmd/gosec@v2.12.0
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.47.2

.PHONY: lint
lint:
	golangci-lint run

.PHONY: sec
sec:
	@gosec -exclude=G401,G204,G505,G101 -quiet ./...

.PHONY: bin
bin:
	@go build -o ./dist/asl

.PHONY: test
test:
	@go test ./... -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: all
all:
	@make -s bin test lint sec

