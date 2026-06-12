.PHONY: all check test lint vet fmt fmt-check build install-tools setup-hooks

GOBIN := $(CURDIR)/bin
GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')
GOLANGCI := $(GOBIN)/golangci-lint

all: check

check:
	@./scripts/check.sh

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

lint: $(GOLANGCI)
	$(GOLANGCI) run ./...

fmt:
	gofmt -s -w $(GOFILES)

fmt-check:
	@test -z "$$(gofmt -l $(GOFILES))" || (gofmt -l $(GOFILES); echo "run: make fmt"; exit 1)

build:
	go build -o bin/rippl ./cmd/rippl

$(GOLANGCI):
	@$(MAKE) install-tools

install-tools:
	@mkdir -p $(GOBIN)
	GOBIN=$(GOBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

setup-hooks:
	git config core.hooksPath .githooks
	@echo "Git hooks path set to .githooks (pre-commit runs: make check)"
