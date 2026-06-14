# Rippl

Go CLI for change impact analysis in Go modules — see which files a change affects, how risky they are, and which tests to run.

![rippl analyze demo](assets/demo.gif)

## Install

Requires [Go 1.22+](https://go.dev/dl/).

```bash
go install github.com/anggasct/rippl/cmd/rippl@latest
```

Or build from source:

```bash
git clone https://github.com/anggasct/rippl.git
cd rippl
go build -o rippl ./cmd/rippl
```

## Quick start

From the root of any Go module:

```bash
# Impact analysis (TUI when stdout is a terminal; use --format text in scripts)
rippl analyze internal/auth/jwt.go

# Risk score breakdown
rippl score internal/auth/jwt.go

# Run tests in packages affected by a change
rippl test internal/auth/jwt.go

# Export the full dependency graph
rippl graph --format mermaid
```

Export formats for `analyze` and `graph`:

```bash
rippl analyze handler.go --format json
rippl analyze handler.go --format mermaid
rippl graph --format json
```

Optional config: `.rippl.yaml` at module root. Graph cache is stored under `.rippl/cache/` — add `.rippl/` to your `.gitignore`.

## Architecture

```mermaid
flowchart TD
    shell[Developer shell]
    cli["cmd/rippl — Cobra CLI"]
    engine[Engine pipeline]
    cache[".rippl/cache/"]

    shell --> cli
    cli --> engine
    engine --> cache

    subgraph engine [Engine]
        parser[parser]
        graph[graph]
        git[git]
        scorer[scorer]
        testmap[testmap]
        impact[impact BFS]
        render[render]
        parser --> graph
        graph --> git
        graph --> scorer
        graph --> testmap
        graph --> impact
        impact --> render
    end
```

Commands: `analyze` | `score` | `test` | `graph`

## Known limits

| Limitation | MVP behavior |
|------------|--------------|
| Dynamic calls / reflection | Not tracked; may miss edges |
| Implicit interface satisfaction | Skipped (planned Phase 2) |
| Generated code | Ignored via config patterns |
| Cross-module internal deps | Module boundary only |

## Releasing

Rippl ships as a Go module — no separate release binaries. Version tags are created automatically by [release-please](https://github.com/googleapis/release-please) via [`.github/workflows/release.yml`](.github/workflows/release.yml).

### Maintainer flow

1. Merge feature/fix PRs to `main` using [Conventional Commits](https://www.conventionalcommits.org/) (`feat(cap-xxx):`, `fix(cap-xxx):`, etc.).
2. CI (`make check`) must pass on each PR.
3. After pushes to `main`, release-please opens or updates a **Release PR** (changelog + version bump).
4. Review and merge the Release PR → GitHub tag `vX.Y.Z` and GitHub Release are created.
5. Users install:

```bash
go install github.com/anggasct/rippl/cmd/rippl@latest
go install github.com/anggasct/rippl/cmd/rippl@v0.1.0
```

### Semver mapping

| Commit type | Version bump |
|-------------|--------------|
| `fix(...):` | patch |
| `feat(...):` | minor |
| `feat(...)!:` or `BREAKING CHANGE:` | major |
| `chore`, `docs`, `test` | no bump |

### First release bootstrap

The initial manifest version is `0.1.0` (see [`.release-please-manifest.json`](.release-please-manifest.json)). The first Release PR may propose `v1.0.0` if releasable `feat` commits since bootstrap warrant a major/minor bump.

**Repo setting (one-time):** In GitHub → Settings → Actions → General → Workflow permissions, enable **Allow GitHub Actions to create and approve pull requests** so release-please can open Release PRs automatically.

`rippl version` prints `dev` when built from source without release ldflags.

## Local development

One-time setup:

```bash
make install-tools   # golangci-lint
make setup-hooks     # pre-commit hook → runs make check-fast (gofmt + vet)
```

Run the full CI-equivalent checks before you push:

```bash
make check           # same as GitHub Actions
```

Pre-commit runs a fast subset (`gofmt` + `go vet`); CI runs the full suite.

Individual targets: `make check-fast`, `make bench-graph`, `make test`, `make lint`, `make vet`, `make fmt`, `make build`.

### Regenerate demo GIF

Requires [vhs](https://github.com/charmbracelet/vhs), `ffmpeg`, and `ttyd`:

```bash
sed "s|REPO_ROOT|$(pwd)|g" assets/demo.tape | vhs -
```

## Documentation

Delivery specs and process live in `project-docs/` (see [AGENTS.md](AGENTS.md)).

## License

MIT — see [LICENSE](LICENSE).
