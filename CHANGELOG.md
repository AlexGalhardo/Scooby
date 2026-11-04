# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Repository standards: `.markdownlint*.jsonc`, `.githooks/commit-msg`, `.config/cliff.toml`, `biome.json`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`, `SECURITY.md`, `LICENSE.md`, issue and pull request templates, `CODEOWNERS`.
- Documentation site in `site/` (Next.js, generated from `docs/`), its Docker image, CI job and the `trading.vukem.dev` Railway service.
- `docs/trading/` knowledge base from ten reference trading projects, `docs/brainstorming.md` owner decisions, setup guides for local Docker, VPS and Railway.
- `AGENTS.md`, `CLAUDE.md`, `.agents/rules` and `.agents/skills`.
- Workflows: `scorecard`, `codeql`, `pr-title`, weekly `vuln`; `.betterleaksignore` with reviewed history findings.
- `vukem-feedback.md` (git-ignored) to collect improvement and bug feedback for the Vukem framework.

### Changed

- The git remote points to `AlexGalhardo/vukem-trading`.
- Frontend tooling moved to Bun 1.4.4: `bun check` replaces `tsc`, Biome 2 lints and formats, the Dockerfile builds assets on `oven/bun:1.4.4-alpine`.
- Documentation, UI text, messages and comments translated to English; documentation files renamed to English names.
- Release workflow publishes the matching CHANGELOG section, a GHCR image and build provenance.
- `Terminal` tables use stable row keys.
