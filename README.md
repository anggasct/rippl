# Rippl

Go CLI for change impact analysis in Go modules.

## Build

```bash
go build -o rippl ./cmd/rippl
```

## Usage

```bash
rippl --help
rippl analyze <file>
rippl score <file>
rippl test <file>
rippl graph
```

Optional config: `.rippl.yaml` at module root. Cache is stored under `.rippl/cache/` — add `.rippl/` to your project's `.gitignore`.

## Local development

One-time setup:

```bash
make install-tools   # golangci-lint
make setup-hooks     # pre-commit hook → runs `make check`
```

Run the same checks as GitHub Actions before you commit:

```bash
make check
```

Individual targets: `make test`, `make lint`, `make vet`, `make fmt`, `make build`.

## Verify

```bash
make check
```

## Documentation

Delivery specs and process live in `project-docs/` (see [AGENTS.md](AGENTS.md)).
