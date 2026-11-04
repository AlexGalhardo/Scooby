# ROADMAP: trading.vukem.dev

**Primary document for onboarding and continuity across sessions and agents.**

Updated: 2026-11-04. Repository: [AlexGalhardo/vukem-trading](https://github.com/AlexGalhardo/vukem-trading). Documentation: <https://trading.vukem.dev>.

Legend: `[x]` done and verified, `[~]` partly done (the note says what is missing), `[ ]` to do. A box is checked only when its **audit assert** passes. Never check a box from a fixture or a document alone.

## 1. Resume point

The personal **paper-only terminal is implemented**: Vukem web, an independent Go engine, backtesting, risk and adapters. Railway runs web, engine and PostgreSQL; the web is at <https://scooby-web-production.up.railway.app> (the service names still say `scooby`).

On 2026-11-04 the project got its repository standards (lint, hooks, community files), moved the frontend to Bun 1.4.4 with `bun check` and Biome, added release and security workflows, and gained a researched knowledge base in [docs/trading/](docs/trading/README.md). Owner decisions are in [docs/brainstorming.md](docs/brainstorming.md).

**Next concrete action:** finish Phase 7 (English-only repository and renames), then Phase 8 (publish the documentation site at `trading.vukem.dev`).

Owner-side items that block validation, not coding: real Google OAuth, Alpaca Paper credentials, licensed B3 intraday data, Telegram alerts and a prolonged paper observation period. Real capital stays out of scope.

## 2. Onboarding for a new session

1. Run `git status --short --branch` and `git log -5 --oneline`; preserve existing changes.
2. Read this file, [AGENTS.md](AGENTS.md), [README.md](README.md) and [docs/brainstorming.md](docs/brainstorming.md).
3. Read the ADRs ([docs/adrs/](docs/adrs/README.md)), especially 0011 to 0014, and the SDD of the task.
4. Read the matching rules in [.agents/rules/](.agents/rules/) before editing.
5. Check the code: a checkbox is not evidence of implementation.
6. Do the smallest authorized delivery, run the focused checks, and update this file with evidence, pending items and the next action.

Do not ask for tokens in chat; use secure provider variables. Never print secrets.

## 3. Scope and decisions

| Topic | State |
| --- | --- |
| Audience | Owner only; one Google account; no public sign-up |
| Money | Paper now, live later behind gates (Phase 17); no live adapter exists |
| Markets | B3 and US equities today; crypto and forex planned; BRL and USD ledgers segregated |
| Execution | Alpaca Paper (US); internal replay (B3); simulator for the rest |
| Strategies | SMA crossover, Z-score reversion, breakout today; opening range, VWAP, ADX trend, grid and DCA planned |
| Data | CSV import; IEX through Alpaca; free APIs, broker paper accounts and paid feeds planned behind provider interfaces |
| Risk, non-negotiable | Mandatory stop on every order; daily-loss kill switch; risk per trade as a percentage of capital; exposure caps per asset and in total |
| Stack | Go 1.27.3, Vukem 0.37.0, PostgreSQL 18, React 19 and Inertia 3, Tailwind 4, Bun 1.4.4, Biome 2 |
| Infra | Railway: web, engine, PostgreSQL, plus the `site` service |
| Docs | Next.js site in `site/`, generated from the Markdown in `docs/` |

## 4. Repository map

| Path | Purpose |
| --- | --- |
| `cmd/app`, `cmd/engine`, `cmd/e2e` | Web server, paper engine worker, end-to-end fixture server |
| `app/`, `routes/`, `bootstrap/` | Vukem application code |
| `internal/trading` | Decimals, sizing, reservations, ledger, calendars, CSV, backtests, store |
| `internal/engine` | Leader election, fencing, replay, dispatch, reconciliation |
| `internal/adapters` | Alpaca paper, data and calendar; Telegram alerts |
| `resources/js` | Inertia React terminal; `components/ui` is generated shadcn/ui |
| `site/` | Documentation site (Next.js); content generated from `docs/` |
| `docs/` | Product, architecture, ADRs, SDDs, setup guides, trading knowledge base |
| `.agents/` | Rules and skills for agents |
| `.github/workflows/` | `ci`, `release`, `scorecard`, `codeql`, `pr-title`, `vuln` |

## 5. Phases

### Phase 0: Foundation and specification

- [x] Official starter, PostgreSQL, Compose, Git and CI.
- [x] Product, architecture, data model, planning, research and onboarding documents.
- [x] Six SDDs and ADRs 0001 to 0014.
- [x] Owner decisions recorded (grilling plus the 2026-11-04 interview).

**Audit assert:** `ls docs/adrs docs/sdd` lists every file linked from `docs/adrs/README.md` and `docs/sdd/README.md`.

### Phase 1: Identity and dashboard

- [x] Google-only login, single account, persistent subject, state, nonce and PKCE.
- [x] Shared sessions and cache, rotation and revocation, CSRF, OAuth throttling.
- [x] Private `/`, `/login`, logout; no local password or sign-up surface.
- [x] Responsive sidebar, theme, charts, empty, error and loading states.
- [x] Server guards on pages and APIs; no credential in props or bundle.
- [ ] Exercise real Google OAuth with the final domain and callback.

**Audit assert:** the e2e login spec passes (`bun run test:e2e`); an unauthenticated request to `/api/snapshot` returns 401.

### Phase 2: Financial data

- [x] Segregated accounts, modes and currencies; instruments with tick and lot.
- [x] Atomic CSV import with provenance, license, hash, OHLCV, offsets and a 5-minute grid.
- [x] Rejection of invalid, duplicate and out-of-order rows; gaps identified.
- [x] IEX history through the adapter; Alpaca calendar and clock.
- [x] B3 2026 calendar, holidays, special opening, session grid changes; unreviewed years fail closed.
- [ ] Receive and validate a licensed B3 intraday dataset.
- [ ] Validate IEX with the chosen account's permissions and limits.

**Audit assert:** `go test ./internal/trading -run 'Calendar|Import|Dataset'` passes; importing a sample with a duplicate row fails.

### Phase 3: Strategies and backtesting

- [x] SMA crossover, mean reversion and breakout, deterministic, with warm-up.
- [x] Immutable configurations, numbered versions, activation separate from the account.
- [x] Persistent, cancelable, recoverable jobs with timeouts.
- [x] Next eligible bar, costs, slippage, participation cap, conservative stop, gap-through.
- [x] Dataset, configuration and code identified; equity curve, benchmark, JSON artifact.
- [x] Return, cost, drawdown, turnover, trades, win and loss metrics; 70/30 partitions.
- [ ] Evaluate strategies on licensed, significant out-of-sample data.

**Audit assert:** two runs with the same dataset hash and configuration produce byte-identical reports.

### Phase 4: Risk, paper and recovery

- [x] Decimal sizing; cash, exposure, daily-loss and position limits.
- [x] Reservation plus intent and outbox in one snapshot and transaction, with an account lock.
- [x] Deduplicated fills, partial fills, costs, ledger and position P&L.
- [x] Independent worker, advisory lock, standby and `leader_epoch`.
- [x] Reconciliation before new entries; timeout without blind retry; `unknown` state.
- [x] Block, cancel and flatten are distinct and audited.
- [x] Tests for concurrent reservations, rollback, timeout, partial, dedupe and failover.
- [ ] Exercise contracts, cancels and fills on a dedicated real Alpaca Paper account.

**Audit assert:** `go test -race ./internal/engine ./internal/trading` passes against a `_test` database.

### Phase 5: Railway and operations

- [x] Dashboard on PostgreSQL snapshots with authenticated polling.
- [x] Audit trail, heartbeat and engine readiness.
- [x] Two binaries, non-root image, Compose; sequential deploy with healthcheck and migrations.
- [x] Audited owner recovery with session revocation; backup and restore rehearsed on a test database.
- [x] Railway project, services, variables and domain created; web and engine healthy.
- [x] Daily PostgreSQL backup scheduled (provider retention six days).
- [x] Deploy on push to `main` verified with the Bun Dockerfile (2026-11-04).
- [ ] Confirm budget and region; add external monitoring.
- [ ] Validate the first backup run and a restore on the provider.
- [ ] Test Google, broker and Telegram integrations in production.

**Audit assert:** `/` returns 302, `/login` 200, `/ready` 200, `/api/snapshot` without a session 401.

### Phase 6: Continuous validation

- [x] Instrumentation and reports for drawdown and cost review.
- [ ] Observe paper trading across the agreed periods; log divergences from the backtest.
- [ ] Compare out-of-sample results without optimizing on the final test set.
- [ ] Measure latency, capacity, RPO and RTO.
- [ ] Record a keep-or-disable decision per strategy.

Phase 6 depends on time, data and configured accounts; do not check it from synthetic fixtures.

### Phase 7: Repository standards and English-only

Goal: the repository follows the ecosystem standards and everything is in English.

- [x] **7.1 Git remote** points to `AlexGalhardo/vukem-trading`. Assert: `git remote -v`.
- [x] **7.2 Community and policy files:** `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`, `SECURITY.md`, `LICENSE.md`, issue and PR templates, `CODEOWNERS`. Assert: files exist and `bun run lint:md` passes on them.
- [x] **7.3 Tooling:** `.markdownlint*.jsonc`, `.githooks/commit-msg`, `.config/cliff.toml`, `biome.json` (v2), `.editorconfig` with tabs. Assert: `bun run lint` and `bun run lint:md` pass.
- [x] **7.4 Bun 1.4.4:** lockfile `bun.lock`, `packageManager`, Dockerfile on `oven/bun:1.4.4-alpine`, `bun check` instead of `tsc`. Assert: `bun run typecheck` prints "No type errors"; `docker build .` succeeds.
- [x] **7.5 Workflows:** `ci`, `release` (CHANGELOG-driven, GHCR image, provenance), `scorecard`, `codeql`, `pr-title`, `vuln`. Assert: YAML parses; CI green on `main`.
- [x] **7.6 Agent files:** `AGENTS.md`, `CLAUDE.md`, `.agents/rules`, `.agents/skills`. Assert: files exist and AGENTS.md links resolve.
- [x] **7.7 Documentation translated** to English with renamed files. Assert: `grep -rP "[áàâãéêíóôõúç]" docs` returns nothing.
- [x] **7.8 Code strings translated** (UI, messages, comments, e2e). Assert: the same grep over `app internal cmd resources/js e2e` returns nothing; `go test ./...` and CI e2e pass (2026-11-04).
- [x] **7.9 Rename `Scooby` to `vukem-trading`:** module path and imports, ldflags, env examples, image tag, comments and docs. Assert: `go build ./...`, `go vet ./...` and `go test ./...` pass; `grep -rI Scooby` finds only the items below.
  - Kept on purpose: the PostgreSQL fixture user, password and database `scooby` (Compose and CI) and the Compose project name, because renaming them orphans local volumes and would change the secret-scanner rule for that fixture value. Rename them together in a dedicated change if wanted.
  - Railway services are already named `web`, `worker` and `site`; the Railway domain keeps its `scooby-web-production` hostname until a custom domain replaces it.
- [x] **7.10 `.golangci.yml`** from the ecosystem template, `golangci-lint` 2.15.0, CI job `Go lint`. Assert: `golangci-lint run ./...` prints `0 issues`. Doc comments on every exported name are off on purpose (`revive` `exported`); package comments stay required.
- [~] **7.11 OpenSSF Scorecard** green. Done: `protect-main` ruleset (no deletion or force-push, linear history, seven required checks, admin bypass), actions pinned by SHA, minimal permissions, SAST (CodeQL), secret scanning, provenance on releases, `re2js` high advisory fixed through an override. Open (see the code-scanning alerts of the Scorecard tool): digest-pin Docker base images (`image:tag@sha256`), the remaining moderate and low advisories (`minimatch` under `shadcn` and `eslint`; wait for upstream or override), fuzz tests for the CSV and decimal parsers, the OpenSSF best-practices badge. Not pursued on purpose: required pull-request review (the owner pushes to `main` directly). Assert: `gh api "repos/AlexGalhardo/vukem-trading/code-scanning/alerts?tool_name=Scorecard&state=open"` lists only the items named here.
- [x] **7.12 First release `v0.1.0`:** CHANGELOG section, annotated tag, GitHub release with the CHANGELOG notes, GHCR image with provenance (workflow run 41920869847 succeeded). Assert: `gh release view v0.1.0` shows the notes.
- [x] **7.13 Pin remaining open ranges** in `package.json` to exact versions (9 dependencies pinned to the installed versions). Assert: `grep -c '"^' package.json site/package.json` prints 0 for both; `bun install --frozen-lockfile` passes.

### Phase 8: Documentation site and DNS

Goal: `trading.vukem.dev` opens the technical documentation, with the same look as vukem.dev.

- [x] **8.1 Site shell** copied from the vukem site and rebranded (`site/`). Assert: `cd site && bun run build`.
- [x] **8.2 Content sync:** `bun run sync:docs` generates `content/docs` and `navigation.ts` from `docs/`. Assert: `bun run check:mdx` passes (54 pages).
- [x] **8.3 Site Dockerfile** (`site/Dockerfile`, context repository root, pinned Bun and Node images, non-root). Assert: `docker build -f site/Dockerfile .` succeeds.
- [x] **8.4 CI job** for the site (lint, `check:mdx`, build). Assert: job green.
- [x] **8.5 Railway `site` service** deployed after green CI. Assert: its Railway URL returns 200 and `/` redirects to `/docs/introduction`.
- [x] **8.6 DNS:** `trading.vukem.dev` CNAME and Railway verification TXT created in the Vercel zone for `vukem.dev` on 2026-11-04; Railway reports the domain verified with a valid certificate. Assert: public resolvers return the Railway edge and `/` answers 307 to `/docs/introduction` over valid TLS.
- [ ] **8.7 Last-updated metadata and sitemap** refreshed (`bun run last-updated`). Assert: `check:last-updated` passes.

### Phase 8b: Known flaky test

- [x] **8b.1 `TestPostgresWorkerFailover`** instability fixed at the root: `Engine.Run` returned the error of an operation interrupted by its own context cancellation (`commit`, `begin`, `leadership connection lost`). Cancellation is now a clean stop. Evidence: the old code failed 20 of 60 runs under `-race`; the new code passes 50 of 50 (`go test -race -count=50 -run 'Failover|CancelledContext' ./tests`), plus a deterministic test for a cancelled context.

### Phase 9: Backtest realism and validation

Reference: [docs/trading/backtesting.md](docs/trading/backtesting.md).

- [ ] **9.1 Parameter declarations as data** (name, type, range, step, default) shared by backtester, UI and optimizer. Assert: a strategy exposes its schema and the UI renders it.
- [ ] **9.2 Look-ahead test:** perturb future candles and assert earlier signals are unchanged, for every strategy. Assert: `go test -run LookAhead ./internal/trading`.
- [ ] **9.3 Partial fills over several bars** with the participation cap; resting-order semantics (price must trade through). Assert: unit tests for a 3-bar partial fill and for a touch without fill.
- [ ] **9.4 Fill-rule ADR** (walk the book versus crossing) written before any broker-style simulator. Assert: ADR merged.
- [ ] **9.5 Walk-forward runner** with windows and a minimum trade count per window. Assert: a report lists per-window metrics.
- [ ] **9.6 Robustness statistics:** block-bootstrap Monte Carlo on trades, cost stress (2x costs and slippage), deflated Sharpe. Assert: report includes the three and flags runs under 30 trades.
- [ ] **9.7 Three-period split** (train, test, validation) with validation touched once per candidate. Assert: the API refuses a second validation run for the same configuration without an override.
- [ ] **9.8 Report metrics:** Sharpe and Sortino with stated risk-free rate, Calmar, SQN, exposure, Buy and Hold baseline. Assert: golden-file test.

### Phase 10: Risk hardening

Reference: [docs/trading/risk-and-execution.md](docs/trading/risk-and-execution.md).

- [x] **10.1 Master arm switch and default deny:** a new account starts blocked (`Halted`), entries are rejected with the reason `account blocked`, exits stay allowed. Assert: `TestDefaultDenyRejectsEntriesButAllowsExits`. Per-venue switches arrive with a second venue.
- [x] **10.2 Mandatory stop** on every entry, enforced inside `State.Reserve` (`MinStopBPS`, reason `stop required`), not only by `Strategy.Validate`. Assert: `TestEntryWithoutStopIsRejected` (0, 9 and -5 bps) and `TestEntryCarriesItsStop`.
- [x] **10.3 Risk per trade in basis points** drives sizing; the loss at the stop, including costs and slippage, never exceeds the budget. Assert: `TestRiskPerTradeSizingInBasisPoints` at 25, 50 and 100 bps.
- [x] **10.4 Daily-loss kill switch:** blocks entries, advances the generation to fence in-flight work, allows exits, lifts at the next market day (`State.NewDay`, only for the daily-loss reason; every other halt stays until the owner clears it), and every halt and resume is audited (`risk.halt`, `risk.resume` in `Store.Mutate`). Assert: `TestDailyLossKillSwitchBlocksEntriesNotExits`, `TestNewDayLiftsOnlyTheDailyLossHalt`.
- [x] **10.5 Protections** (`Policy`: `stop_streak`, `streak_lock_mins`, `cooldown_mins`, `drawdown_bps`; zero disables, new accounts default to a 3-stop streak with a 240-minute lock and a 10% drawdown halt): stop-loss streak lock, per-symbol cooldown, equity-drawdown halt. Exposed in the risk form. Assert: `TestStopStreakLocksTheSymbolThenExpires`, `TestATargetExitResetsTheStopStreak`, `TestCooldownAppliesAfterAnyExit`, `TestEquityDrawdownHaltsTheAccount`. Shortcut: a streak counts the last fill of the closing exit, not the whole position lifecycle.
- [x] **10.6 Exit model:** every exit order carries `close_reason` (`stop`, `target`, `session_close`, `signal`, `risk`, `manual`); the engine sets it for stop, target and session exits. The time-limit barrier stays with Phase 12 strategy work. Assert: `TestEveryExitCarriesACloseReason`.
- [x] **10.7 Alarms:** `State.Watch` alerts once per condition (an order `unknown` for 10 minutes or more; any halt such as the kill switch or a stale feed) and the engine sends them through Telegram when configured. Assert: `TestWatchAlertsOncePerCondition`.

### Phase 11: Data providers

Reference: [docs/trading/openbb.md](docs/trading/openbb.md) and [docs/trading/strategies-and-markets.md](docs/trading/strategies-and-markets.md).

- [ ] **11.1 Provider interface** (`transform_query`, `extract`, `transform` pattern in Go) with credentials validated before the call. Assert: a provider with a missing key fails before any network call.
- [ ] **11.2 Coverage matrix** per market and data mode, in the API and the UI. Assert: snapshot includes live, delayed, end-of-day or none per market.
- [ ] **11.3 Crypto candles** (public exchange API) and a testnet paper adapter for one venue. Assert: backtest on a downloaded BTC and USDT dataset reproduces.
- [ ] **11.4 Forex dataset import** (backtest only) with provenance. Assert: import test with a sample file.
- [ ] **11.5 Source terms** recorded per provider in [docs/sources.md](docs/sources.md). Assert: every provider in code has a row.

### Phase 12: Strategy expansion

- [ ] **12.1 Opening-range breakout** with session-aware bounds. Assert: backtest on the intraday sample runs and passes 9.2.
- [ ] **12.2 VWAP reversion.** Assert: same.
- [ ] **12.3 ADX-filtered trend.** Assert: same.
- [ ] **12.4 Grid and DCA (spot)** with capped inventory and a range exit. Assert: inventory never exceeds the configured cap in a trending fixture.
- [ ] **12.5 Market making** stays out of scope until order-book data and an accepted fill-rule ADR exist.

### Phase 13: Terminal UX

Reference: [docs/trading/terminal-and-agents.md](docs/trading/terminal-and-agents.md).

- [ ] **13.1 Safety bar:** arm switch, kill switch, account state banner, live risk panel. Assert: e2e asserts the banner changes when the account is blocked.
- [ ] **13.2 Command palette and numbered shortcuts.** Assert: e2e opens a page through the palette.
- [ ] **13.3 Saved workspace layout** per user. Assert: layout survives a reload.
- [ ] **13.4 Watchlist** with a size cap and a coverage-matrix view.
- [ ] **13.5 Paper hotkey order templates** sized by percentage of buying power. Assert: e2e places a paper order with a hotkey.
- [ ] **13.6 Canvas tape and ladder** for streamed data. Assert: no React state updates per tick (profiling note in the PR).
- [ ] **13.7 Generated TypeScript wire types** from Go structs with a drift check in CI. Assert: CI fails when a Go struct changes without regenerating.
- [ ] **13.8 Alerts:** thresholds with a checker interval and Telegram delivery. Assert: fault-injection test.

### Phase 14: Observability and operations

- [ ] **14.1 Metrics:** heartbeat age, feed staleness, reconciliation drift, rejected-order rate, slippage versus expectation. Assert: exposed on an authenticated endpoint.
- [ ] **14.2 External uptime check** on `/up` and the site. Assert: alert fires on a forced failure.
- [ ] **14.3 Runbooks** for incident, restore and key rotation in `docs/`. Assert: restore rehearsal log committed.
- [ ] **14.4 Dependency policy:** Dependabot, `govulncheck`, OSV-Scanner on the site, betterleaks. Assert: all jobs green on `main`.

### Phase 15: Agents and MCP (read-only)

- [ ] **15.1 Read-only MCP server** for market data and reports with a bearer token and a tool allowlist. Assert: a request without a token is rejected.
- [ ] **15.2 No order tools** are exposed to any agent. Assert: tool list test.
- [ ] **15.3 LLM-assisted strategy drafting** produces a validated configuration for human review only. Assert: the draft cannot be activated without an owner action.

### Phase 16: Multi-user (explicitly out of scope)

Not planned. Reopen only with a new ADR covering isolation, authorization, billing and regulation.

### Phase 17: Live-readiness gate (decision only)

No live adapter is built in this roadmap. Writing an ADR is the only allowed work until all of these hold:

- [ ] Paper results over an agreed period match backtest expectations within a stated tolerance.
- [ ] Risk gate, kill switch and reconciliation have failure-path tests in CI.
- [ ] Live is enabled by a server-side flag separate from any agent or UI setting.
- [ ] Credentials encrypted at rest and present only in the worker.
- [ ] First-period limits (order size, daily loss) set deliberately low.
- [ ] Owner sign-off recorded in an ADR.

## 6. Contracts and limitations

- Financial persistence uses versioned decimal JSON in a PostgreSQL account with lock and compare-and-swap plus audit in the same commit; do not claim normalized `NUMERIC` tables that do not exist. ADR-0012 explains the trade-offs and the migration path for high volume.
- The outbox is the durable list of orders approved in the snapshot. There is no external exactly-once; an in-flight request needs reconciliation even with generation and fencing.
- Web polling is 10 s, the execution loop 15 s, candles 5 m. This is not HFT, and stops are not guaranteed broker-native stops.
- The B3 calendar supports 2026 only. Extend a year only with a reviewed calendar and tests. IEX is not the consolidated market.
- The financial snapshot grows with history. Never truncate the ledger to save space; normalize or archive with a trail first.
- One active configuration per US asset and one active B3 replay per account.
- B3 is an internal historical replay, not a broker or a live feed.
- The web may use paper keys to import IEX data; all values stay on the server. Never place credentials in `VITE_*` variables, props, logs or Git.
- Push to `main` deploys to Railway. Use pull requests with green CI.
