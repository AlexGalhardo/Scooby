---
paths:
  - "CHANGELOG.md"
---

# Changelog rules

- Keep a Changelog 1.1.0 and SemVer tags `vX.Y.Z`. The format is configured in `.config/cliff.toml`.
- Add entries under `[Unreleased]` in the same change that alters behavior, commands or conventions, grouped as Added, Changed, Fixed, Removed (and Security, Deprecated when needed).
- Releasing: move `[Unreleased]` into a `## [X.Y.Z] - YYYY-MM-DD` section, update the compare links at the bottom, commit `chore(release): vX.Y.Z`, create the annotated tag `vX.Y.Z` on `main` and push the commit and the tag. The release workflow publishes the matching CHANGELOG section as the GitHub release notes and refuses a tag without one.
- Push version tags one at a time. GitHub creates no push event when more than three tags go up in one push. For a tag already pushed without a release, run the `Release` workflow manually with the version.
- Never move a pushed tag.
