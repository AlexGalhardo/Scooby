# ADR-0017: documentation site generated from `docs/`

State: implemented on 2026-11-04 by owner instruction.

## Context

The owner wants <https://trading.vukem.dev> to open the technical documentation with the same look
as vukem.dev. The repository documentation is plain Markdown written for GitHub and must stay the
single source of truth.

## Decision

- `site/` is the Next.js documentation site of the Vukem ecosystem (App Router, Tailwind, MDX),
  copied and rebranded. Its stack versions are the ones of the reference site, pinned exactly.
- `site/scripts/sync-docs.ts` generates `site/content/docs/**` and `navigation.ts` from `docs/`, the
  root community files and `CHANGELOG.md`. CI regenerates and fails on any difference.
- The site is a separate Railway service built from `site/Dockerfile` with the repository root as
  the build context. `/` redirects to `/docs/introduction`.
- DNS: a `CNAME` for `trading` and the Railway verification `TXT` record in the `vukem.dev` zone.

## Alternatives

- Vercel hosting: the reference site uses it, but the project already runs on Railway.
- Maintain MDX by hand: duplicates the repository documentation and drifts.

## Consequences

Documentation edits land in `docs/` and are published by the next deploy. The converter handles
Markdown that MDX rejects (autolinks, braces, relative links); a new construct may need an addition.

## Review trigger

The reference site changes its content model, or the converter needs more than a few rules.
