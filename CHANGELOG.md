# Changelog

## [0.1.1](https://github.com/anggasct/rippl/compare/v0.1.0...v0.1.1) (2026-06-14)


### Bug Fixes

* dogfood improvements (analyze, TUI, score, docs) ([#17](https://github.com/anggasct/rippl/issues/17)) ([fc8f42b](https://github.com/anggasct/rippl/commit/fc8f42b27e3a4a6be72dfe489daa92ef944198e5))

## 0.1.0 (2026-06-14)

### Features

- CLI commands: `analyze`, `score`, `test`, `graph`
- Interactive TUI and text, JSON, and Mermaid export
- File-level dependency graph with cached builds
- Impact traversal (BFS) with risk scoring and git history signals
- Affected test package resolution
- Zero-config for standard Go modules (optional `.rippl.yaml`)
