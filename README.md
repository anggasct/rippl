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
go install github.com/anggasct/rippl/cmd/rippl@v1.0.0
```

### Semver mapping

| Commit type | Version bump |
|-------------|--------------|
| `fix(...):` | patch |
| `feat(...):` | minor |
| `feat(...)!:` or `BREAKING CHANGE:` | major |
| `chore`, `docs`, `test` | no bump |

### First release bootstrap

`v1.0.0` is the initial tagged release (see [CHANGELOG.md](CHANGELOG.md)).

### Troubleshooting release-please

**`GitHub Actions is not permitted to create or approve pull requests`**

One-time repo setting (repo admin):

1. GitHub → **Settings** → **Actions** → **General**
2. Under **Workflow permissions**, choose **Read and write permissions**
3. Enable **Allow GitHub Actions to create and approve pull requests**
4. Re-run the failed **Release** workflow (or push a commit to `main`)

**`Pull request body did not match` / release PR merged but no tag**

Release PRs must be opened by release-please (bot), not created manually. If a Release PR was merged without the bot footer, create the tag manually:

```bash
gh release create vX.Y.Z --target <sha> --title "vX.Y.Z" --notes-file CHANGELOG.md
```

After `v1.0.0` exists, future `feat`/`fix` merges to `main` should only bump via bot-opened Release PRs.

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
