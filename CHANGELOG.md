# Changelog

## [0.3.0](https://github.com/anggasct/rippl/compare/v0.2.0...v0.3.0) (2026-06-14)


### Features

* agent integration (JSON export, diff, context, config example) ([#21](https://github.com/anggasct/rippl/issues/21)) ([89ebce3](https://github.com/anggasct/rippl/commit/89ebce3c93e660d5d9a2ba3afee635f95fd98413))

## [0.2.0](https://github.com/anggasct/rippl/compare/v0.1.1...v0.2.0) (2026-06-14)


### Features

* **cap-200:** add suggested_actions JSON and analyze filter flags ([#19](https://github.com/anggasct/rippl/issues/19)) ([d04f65d](https://github.com/anggasct/rippl/commit/d04f65dbd8eb6f0deb9b68aca657e0b4cbaef413))

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
