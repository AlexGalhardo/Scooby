---
paths:
  - "internal/trading/**"
  - "internal/engine/**"
  - "internal/adapters/**"
---

# Money, trading and risk rules

- Money, prices and quantities use decimal arithmetic (the `Decimal` type in `internal/trading`), never `float64`. Persist and transport them as decimal strings.
- Respect instrument tick size and lot size when sizing and pricing; round in the direction that does not increase risk.
- BRL and USD ledgers stay segregated per account. No cross-currency sums without an explicit FX policy recorded in an ADR.
- Paper only: this project never sends real orders. A live adapter needs an ADR that meets the gates in `docs/trading/risk-and-execution.md`.
- Every entry order carries a stop. The risk engine rejects an order without one.
- Risk order of evaluation: arm switch, instrument and session rules, order-level rules, account caps, loss controls, protections, then any model or AI gate. A model can only reject entries and never blocks an exit.
- Persist the decision, reservation and intent before any external call. A timeout leaves the order `unknown` and is reconciled before any retry; never resend blindly.
- Fills are idempotent by broker fill id; ledger entries are append-only.
- Strategies are deterministic functions of closed candles and parameters, with declared warm-up. Parameter sets are versioned and immutable once used.
- Backtests state their fill rules, costs and slippage, and never present replay as broker execution or promise profitability.
- Pin units (basis points, percent, decimal) in names and tests.
