# vukem-trading documentation site

The public docs at <https://trading.vukem.dev>: Astro 7 (static output), React islands, Tailwind CSS v4, Base UI and MDX, mirroring the `vukem/site` reference. See [ADR-0017](../docs/adrs/0017-documentation-site.md).

## Run

```sh
cd site
bun install --frozen-lockfile
bun run sync:docs        # regenerate content/ from ../docs (commit the result)
bun run dev              # http://localhost:3000
bun run build            # static build into out/
bun run lint
bun run check:mdx        # compile MDX, check components, /docs links and code line length
bun run last-updated     # refresh content/last-updated.json (needs full git history)
bun run check:last-updated
bun run test:scripts
```

Never edit `content/` by hand: it is generated from `docs/`, the root community files and `CHANGELOG.md`. Add a page by adding the Markdown file under `docs/` and listing it in `scripts/sync-docs.ts`.

## Container

`site/Dockerfile` (build context: the repository root) builds `out/` and serves it with nginx on `$PORT`. Railway: `RAILWAY_DOCKERFILE_PATH=site/Dockerfile`, no start command.
