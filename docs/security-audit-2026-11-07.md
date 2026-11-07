# Security audit, 2026-11-07

Full audit of `vukem-trading` at branch `feat/todo-2026-11` (PR 16), using the vendored Cloudflare `security-audit` skill. Machine-readable records: [findings](security-audit-2026-11-07.findings.json) and [coverage ledger](security-audit-2026-11-07.coverage.json), both accepted by the skill validators.

## Scope and method

Profile `standard`, single auditor. Scope: Go web app and engine (Google owner gating, MCP endpoint, broker and Telegram adapters, paper and live gates, order and risk paths, migrations, jobs), `resources/js`, the Astro site, CI workflows, Dockerfiles and Railway configuration. Skipped: vendored skills, `node_modules`, generated output.

Deviations from the skill, stated plainly:

- Reconnaissance, hunting and verification were done by one agent reading the source (about 5,000 lines of Go plus workflows and configs), not by parallel hunter waves. There is no prior ledger, so this is the first pass and does not claim exhaustive coverage.
- No target code was run during the audit; the `govulncheck` and `bun audit` scans found nothing reachable. Each confirmed finding gets a failing regression test before its fix.
- The skill validators need `O_NOFOLLOW`, which Windows lacks. They were run through a preload that only supplies those flag constants; both passed.

## Summary

| Severity | Count | Findings |
| --- | --- | --- |
| Critical | 0 | none |
| High | 0 | none |
| Medium | 1 | TR-AUDIT-001 |
| Low | 2 | TR-AUDIT-002, TR-AUDIT-003 |
| Needs validation | 1 | TR-AUDIT-004 |
| Rejected | 7 | TR-AUDIT-R01 to R07 |

No path was found that places a real order or bypasses the live-readiness gate: the broker host is the hardcoded Alpaca paper endpoint, there is no live adapter, and the gate reads only the process environment and is denied for the web role.

## TR-AUDIT-001: ticket orders use an arbitrarily old price (medium)

- CWE: CWE-1288 (improper validation of consistency within input), CWE-754.
- Location: `app/controllers/terminal/terminal.go:335`, `internal/trading/preferences.go:196`, compare `internal/engine/engine.go:359`.
- Impact: a hotkey ticket on the US account prices and sizes a paper broker order from the last close of the newest stored dataset, however old. The engine halts on quotes older than `Policy.MaxAgeSeconds`; the ticket path skips that gate, so the stale-feed fail-closed rule does not hold for owner orders.
- Reproduction: store an AAPL dataset whose last candle is days old, arm the US account, `POST /api/accounts/us/ticket {"key":"b","symbol":"AAPL"}`. Before the fix the response is 201 with an approved buy.
- Fix: return the candle time with the price and reject a US ticket older than `MaxAgeSeconds` plus one bar. B3 is a replay on a historical clock and is exempt.

## TR-AUDIT-002: MCP endpoint has no throttle or token-strength check (low)

- CWE: CWE-307, CWE-521.
- Location: `routes/web.go:63`, `app/controllers/terminal/mcp.go:68`.
- Impact: `/mcp` is public and protected by one static bearer token. Guesses are never limited and a short `MCP_TOKEN` boots without complaint, which breaks the rule that agent surfaces are rate-limited and that weakening configuration is a boot error. Exposure is read-only metrics.
- Reproduction: send repeated `Authorization: Bearer <guess>` requests; every one is evaluated.
- Fix: per-IP limiter on `/mcp`; refuse to register the route when `MCP_TOKEN` is set and shorter than 32 characters.

## TR-AUDIT-003: MCP tool errors leak internal text (low)

- CWE: CWE-209.
- Location: `app/controllers/terminal/mcp.go:109`, source at `:136`.
- Impact: a store failure inside `tools/call` returns the driver message to the token holder, against the rule that client errors never include SQL.
- Reproduction: make the datasets table unavailable in a test database and call `list_datasets`.
- Fix: return the message only for deliberate `errs` errors, a generic text otherwise.

## TR-AUDIT-004: limiter and HSTS may key on the Railway proxy (needs validation)

`routes/social.go:29` limits Google OAuth starts to 20 per minute per `c.IP()`. The Railway env examples set neither `TRUSTED_PROXIES` nor `TRUST_FORWARDED_PROTO`, so behind the platform proxy all clients may share one bucket (an anonymous caller could keep the owner from starting a login) and HSTS may not be sent. Whether that happens, and the right CIDR, are deployment facts outside the repository. Owner action: read the client IP in the web service log, then set both variables. No code change was made.

## Rejected candidates

| Id | Claim | Why rejected |
| --- | --- | --- |
| R01 | Owner bootstrap race | Row lock, conditional update and subject pinning; tests cover foreign and unverified accounts |
| R02 | Live orders through the Alpaca adapter | Paper host is a constant; no redirects; no live adapter |
| R03 | Unbounded datasets exhaust memory | Only the single owner can create them (self-impact) |
| R04 | `set:html` on the docs site | Repository markdown rendered at build time |
| R05 | Workflow expression injection | Inputs go through env variables; `pr-title` checks out nothing |
| R06 | Telegram token in errors | All errors are generic |
| R07 | Dead `PostController` routable | Not registered; `/posts` is tested as 404 |

## Coverage statement

Covered by source review: authentication and owner gating, every route's auth middleware, MCP, order and risk paths, broker and Telegram adapters, CSV import and SQL, migrations, the React terminal, the docs site, all eight workflows, the Dockerfiles and Railway configs. Gaps: no parallel second pass, no execution of target code before the fixes, and the Railway runtime facts in TR-AUDIT-004.
