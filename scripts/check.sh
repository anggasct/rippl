#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

GOFILES="$(find . -name '*.go' -not -path './vendor/*')"

log() {
	printf '==> %s\n' "$*"
}

fail() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

golangci_lint() {
	if [[ -x "$ROOT/bin/golangci-lint" ]]; then
		"$ROOT/bin/golangci-lint" "$@"
		return
	fi

	if command -v golangci-lint >/dev/null 2>&1; then
		command golangci-lint "$@"
		return
	fi

	local gopath_bin
	gopath_bin="$(go env GOPATH)/bin/golangci-lint"
	if [[ -x "$gopath_bin" ]]; then
		"$gopath_bin" "$@"
		return
	fi

	fail "golangci-lint not found; run: make install-tools"
}

log "gofmt"
unformatted="$(gofmt -l $GOFILES)"
if [[ -n "$unformatted" ]]; then
	printf 'These files need gofmt:\n%s\n' "$unformatted" >&2
	fail "run: make fmt"
fi

log "go vet"
go vet ./...

log "go test (race)"
go test -race -count=1 ./...

log "golangci-lint"
golangci_lint run ./...

log "build"
go build -o /tmp/rippl ./cmd/rippl

log "all checks passed"
