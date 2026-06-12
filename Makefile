.PHONY: all check check-fast test lint vet fmt fmt-check build install-tools setup-hooks bench-graph

GOBIN := $(CURDIR)/bin
GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')
GOLANGCI := $(GOBIN)/golangci-lint

all: check

check:
	@./scripts/check.sh

# Fast subset for pre-commit; full make check runs in CI.
check-fast: fmt-check vet

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

bench-graph:
	go test ./internal/graph/... -bench=BenchmarkLoadOrBuild -count=1 -run=^$$

$(GOLANGCI):
	@$(MAKE) install-tools

install-tools:
	@mkdir -p $(GOBIN)
	GOBIN=$(GOBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

setup-hooks:
	git config core.hooksPath .githooks
	@echo "Git hooks path set to .githooks (pre-commit runs: make check-fast)"
