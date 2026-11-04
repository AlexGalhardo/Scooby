# Brainstorming and owner decisions

Status: living record. Decisions marked **Decided** come from the owner's interview on 2026-11-04 and the earlier grilling in [grilling.md](grilling.md); they are final until the owner changes them. Research behind the options lives in [trading/](trading/README.md).

## Owner decisions (2026-11-04)

| Topic | Decision |
| --- | --- |
| Money | **Paper now, live later** behind safety gates (risk limits, kill switch, manual approval). No real orders in the current scope. |
| Markets | B3, US equities and ETFs, crypto and forex. B3 and US exist today; crypto and forex are roadmap items. |
| Strategies | Trend and momentum, mean reversion, intraday (opening-range breakout and similar), grid, DCA and market making. Market making waits for order-book data. |
| Users | **Only the owner**: personal terminal, single Google account. Multi-user is out of scope. |
| Data | CSV import (exists), free APIs, broker paper accounts (Alpaca, Binance testnet) and paid feeds, all behind provider interfaces. |
| Risk, non-negotiable | Mandatory stop on every order, daily-loss kill switch, risk per trade as a percentage of capital, exposure caps per asset and in total. |
| Work style | Sequential and inline; the `researcher` agent for external research only; no worktrees or parallel agents. |
| Repository | `AlexGalhardo/vukem-trading`; the agent may push to `main`. |
| Deployment | A separate `site` service on Railway, deployed after green CI; DNS record shown before it is applied. |
| Frontend tooling | Bun 1.4.4 with `bun check` for types (no `tsc`), Biome 2 for lint and format. |
| Dependencies | The `site` copies the stack of the vukem site with exact versions. |

## Product alternatives (history)

| Path | Benefit | Cost | Verdict |
| --- | --- | --- | --- |
| Personal terminal | One identity and account | Allowlist, owner recovery | **Chosen** |
| SaaS | Many users | Isolation, billing, support, regulation | Out of scope |
| Research platform | Learn without execution | Differs from market behavior | First increment, done |
| Terminal with paper broker | Tests execution and recovery | Credentials, data plan | Current |

## Architecture alternatives (history)

- **Modular monolith plus a separate worker:** shared Go domain, distinct web and engine lifecycles and credentials. **Chosen**.
- **Microservices per engine:** more network and deployment cost; deferred until load or team justifies it.
- **PostgreSQL events plus outbox:** simple local transactions and recovery. **Chosen**.
- **NATS JetStream:** useful for many consumers; adds a broker to operate. Delivery does not remove the need for idempotency at the broker.
- **Redis:** fine for cache and sessions, never the system of record for orders.
- **TimescaleDB:** evaluate Railway support before depending on it; plain PostgreSQL with indexes and partitions first.

## Open questions and my recommendations

1. **Order-book data for market making.** No source yet. Recommendation: do not start market making; grid and DCA first with capped inventory.
2. **Crypto venue.** Recommendation: Binance testnet for paper, public candles for backtests; one pair family first (BTC and ETH against USDT).
3. **Forex source.** Recommendation: backtest-only with licensed or public daily and hourly data; no broker adapter until there is a paper account to test.
4. **B3 intraday licensing.** Still the main blocker for validating B3 strategies; keep the CSV import path and record provenance and license.
5. **Fill model for passive orders.** Needs an ADR: walk-the-book versus price-crossing. Reference projects disagree or do not document it.
6. **Parameter declarations as data.** Recommendation: adopt now so that the optimizer, the UI and the backtester share a single source.
7. **Provider interface.** Recommendation: define a Go interface per data type (quote, bar, fundamentals) with credentials validated before use.
8. **Live gate criteria.** The list in [trading/risk-and-execution.md](trading/risk-and-execution.md) becomes an ADR before any live adapter.
9. **Tooling migration.** npm to Bun changes the lockfile, the Dockerfile and CI; recorded in ADR-0015.

## Quantitative hypotheses

- Moving averages: explicit entry and exit rules with a volatility-scaled stop.
- Reversion: Z-score with a minimum window and a guard for zero deviation; thresholds are not investment recommendations.
- Breakout: use the previous range and volatility, never the same candle's future high.
- Every strategy defines sizing, exit, session and missing-data behavior. A good in-sample result does not validate a strategy, and parameters are never chosen on the held-out set.

## Not to adopt

An LLM never signs or sends orders; if added, it drafts validated configuration for human review and never runs arbitrary code. Never promise detection of tops or bottoms, exactly-once delivery across a network, guaranteed stop prices or any return.

## Questions for future experiments

Does the feed cover the desired assets? Does the budget cover history and licenses? How do backtest, internal simulation and broker paper differ for each strategy? In which scenarios must a strategy be disabled? Which metrics actually help the owner decide, rather than decorating the dashboard?
