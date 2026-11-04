# Proposed data model

> Implementation update: see the [ROADMAP](../ROADMAP.md),
> [ADR-0011](adrs/0011-autonomous-implementation.md),
> [ADR-0012](adrs/0012-transactional-snapshots.md) and
> [ADR-0013](adrs/0013-polling-workers-and-deploy.md) for the effective state.
> The text below preserves the original requirements/designs, not a statement that
> all infrastructure or external validation details have already occurred.

Status: logical design, no migrations implemented. Audience and market may still
change cardinalities and precision.

## Entities

| Table | Main fields and constraints |
| --- | --- |
| `users` | id, name, verified email, status, timestamps |
| `social_accounts` | user_id, provider, subject; UNIQUE(provider, subject) |
| `sessions` | revocable session, expiration, user; without unnecessary Google token |
| `trading_accounts` | owner_id, broker, external_id, mode, currency, status |
| `instruments` | market, symbol, currency, tick_size, qty_step, timezone, status |
| `candles` | instrument_id, feed, timeframe, open_time, OHLCV, closed, revision |
| `strategy_versions` | strategy_id, version, schema_version, JSONB parameters, hash |
| `strategy_runs` | account_id, version, mode, status, start/end and dataset |
| `signals` | run_id, candle/event, direction, reason; dedupe_key UNIQUE |
| `risk_policy_versions` | account, version, currency, limits and validity |
| `risk_decisions` | signal, version, snapshot, reason, approved/rejected |
| `risk_reservations` | account, intent, reserved exposure, status |
| `order_intents` | account, signal, idempotency_key, client_order_id, state |
| `broker_orders` | account, broker_order_id UNIQUE per account, intent, status |
| `fills` | account, broker_execution_id UNIQUE per account, qty, price, fees |
| `ledger_entries` | cash/position entries, referenced fill and reversal |
| `position_snapshots` | account/instrument, qty, cost and mark; recoverable projection |
| `outbox_events` | type, version, payload, attempts, available_at, claim and generation |
| `engine_controls` | account, halted, reason, generation, heartbeat and leader |
| `reconciliation_runs` | account, result, divergences, cursor and time |
| `backtest_runs` | code, dataset/hash, parameters, seeds, costs and metrics |
| `audit_events` | actor, account, action, previous/new version, request_id and time |

## Precision and units

Proposed: PostgreSQL `NUMERIC` and explicit decimal in Go; decimal strings in the JSON
contracts and display conversion in React. Final scales depend on the
instruments; do not assume two decimal places for price, quantity or crypto. Round
quantity down according to lot/step when opening risk, price according to tick and
side; validate the broker's minimums and maximums. Do not use floats for ledger or limits.
Statistical indicators may use float with tolerances and finite inputs.

Every amount carries a currency. Do not add USD and BRL without an FX rate, instant and FX source.
Without a reliable source, show values separately. Store UTC (`timestamptz`),
keep the market timezone and use the official calendar for holidays and DST.

## Integrity and concurrency

- FKs preserve account/execution/version; external IDs are unique in the correct scope.
- Candles: UNIQUE(instrument_id, feed, timeframe, open_time, revision); an explicit current
  revision prevents mixing revisions in the same backtest.
- Risk reservations and intents are created atomically with the outbox. Include partially filled orders and orders in unknown state in the conservative calculation.
- Fill events are deduplicated before changing the ledger. Corrections are adjustment/reversal entries, never silent deletion of financial evidence.
- Real and paper accounts, synthetic and historical datasets use explicit namespaces.
- Indexes by account+status+time on orders and account+instrument on positions.
- Retention/partitioning of candles only after budget and volume; do not discard
  data that supports results without recording the dataset's unavailability.

## Risk and sizing

For a purchase without leverage: `budget = equity * risk_fraction`;
`risk_per_unit = entry - stop + estimated_costs_per_unit + slippage_buffer`;
`qty = floor_to_step(min(budget/risk_per_unit, cash_limit/entry, exposure_limit/entry))`.
Reject stop >= entry, negative costs, non-positive denominator, non-finite values or
quantity below the minimum. The reservation uses the conservative execution scenario,
not the expected profit. Stops do not limit maximum guaranteed loss.

Realized P&L comes from fills, cost policy and fees; unrealized requires a recent marked
price and source. Result before costs, after costs and tax estimate are distinct
fields; tax calculation is not defined in the MVP.

## Confirmed adjustment: two markets, personal use

The model must distinguish Brazil and the US and preserve separate paper accounts. An
owner may have several `trading_accounts`; `owner_id` serves authorization,
without introducing organizations, billing or SaaS tenancy in the MVP. Instrument
uniqueness uses market/venue/symbol, not the isolated symbol. Separate BRL and USD in
the ledger; FX policy and aggregated limits still depend on a decision. Synthetic data
needs a visible flag and source, including in the internal simulator.
