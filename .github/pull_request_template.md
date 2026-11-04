## Summary

<!-- What does this change and why? End with the roadmap task, e.g. "Refs: P1.2.3". -->

## Checklist

- [ ] PR title is a Conventional Commit (`type(scope): description`, `!` for a breaking change)
- [ ] Focused checks pass locally: `go test -race` for the touched packages, `go vet ./...`, `npm run typecheck`, `npm run lint:md`
- [ ] Any change to risk, execution or ledger logic has a test, including the failure path
- [ ] No credential, token or account identifier in code, logs, props, fixtures or docs
- [ ] No new dependency without an ADR; versions pinned exactly (no `latest` or prereleases)
- [ ] Workflows: actions pinned by full SHA with a `# vX.Y.Z` comment; minimal `permissions`
- [ ] `CHANGELOG.md` updated under `[Unreleased]` when behavior, commands or conventions change
- [ ] Docs, SDDs and ADRs updated in the same change; everything in the repository is in English
