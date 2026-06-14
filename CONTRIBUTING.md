# Contributing

Thanks for your interest in Rippl. This document is for people working on the repository.

## Development setup

```bash
make install-tools   # golangci-lint
make setup-hooks     # optional pre-commit hook
make check           # full CI-equivalent checks
```

Other useful targets: `make test`, `make lint`, `make build`.

### Regenerate demo GIF

Requires [vhs](https://github.com/charmbracelet/vhs), `ffmpeg`, and `ttyd`:

```bash
sed "s|REPO_ROOT|$(pwd)|g" assets/demo.tape | vhs -
```

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/) with **user-facing scopes** so the public changelog stays readable:

```text
feat(analyze): add JSON export
fix(graph): handle empty module
```

Good scopes: `cli`, `analyze`, `score`, `test`, `graph`, `engine`, `docs`.

Put internal capability IDs (for example `CAP-203`) in the **PR description** or commit body — not in the commit scope. Scopes appear in [CHANGELOG.md](CHANGELOG.md) when releases are cut.

## Pull requests

1. Open a PR against `main`.
2. Ensure CI passes (`make check` locally).
3. Get review and merge.

## Releases

Rippl is distributed as a Go module (`go install`). Version tags and [CHANGELOG.md](CHANGELOG.md) updates are published from `main` via the Release workflow in [`.github/workflows/release.yml`](.github/workflows/release.yml).
