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

Rippl is distributed as a Go module (`go install`). Tags are managed by [release-please](https://github.com/googleapis/release-please) via [`.github/workflows/release.yml`](.github/workflows/release.yml).

After merges to `main`, release-please opens a **Release PR** with changelog and version bumps. Merge that PR to publish `vX.Y.Z`.

### One-time GitHub setting

If the Release workflow fails with *GitHub Actions is not permitted to create or approve pull requests*:

1. **Settings** → **Actions** → **General**
2. **Workflow permissions** → **Read and write permissions**
3. Enable **Allow GitHub Actions to create and approve pull requests**

Release PRs must be opened by the release-please bot. Do not create them manually.

### Internal project docs

Maintainers with the Obsidian vault symlink may use `project-docs/` and [AGENTS.md](AGENTS.md) for capability specs and delivery process. That material is not required for general contributors.
