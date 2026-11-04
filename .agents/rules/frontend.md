---
paths:
  - "resources/**"
  - "e2e/**"
---

# Frontend rules

- React 19 and TypeScript (strict), pages under `resources/js/pages`, rendered by Inertia. Prop shapes live in `resources/js/types.ts` and mirror the Go controllers in `app/controllers`.
- UI: shadcn/ui on Radix primitives in `resources/js/components/ui` (generated; excluded from Biome), Tailwind v4 CSS-first with tokens in `@theme`.
- Package manager and runtime: Bun 1.4.4. Types are checked with `bun check` (`bun run typecheck`); lint and format with Biome (`bun run lint`). Both must pass before a task is done.
- All UI text is English.
- Every number shows its currency and the freshness of its data. Never sum BRL and USD.
- Prices, quantities and money arrive as decimal strings; format with the shared decimal helpers, never `Number` arithmetic for ledger values.
- Charts, tape and ladders draw on canvas and never push high-frequency data through React state.
- Never place credentials in props, `VITE_*` variables, bundles, logs or fixtures.
- Accessibility: visible labels, `aria-label` on icon-only buttons, visible focus, `prefers-reduced-motion` respected, no `dangerouslySetInnerHTML` with untrusted content.
- Validate input client-side for UX only; the server is the source of truth.
- For visual changes, verify in a browser (Playwright) and load the `frontend-design` and `web-design-guidelines` skills first.
