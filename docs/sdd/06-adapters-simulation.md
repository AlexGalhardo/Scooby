# SDD-06 — confirmed adapters and simulation modes

> Current implementation and explicit deviations are in the [ROADMAP](../../ROADMAP.md)
> and [ADRs 0011–0013](../adrs/README.md). Local technical acceptance is distinct
> from real OAuth/broker, Railway publication and prolonged financial observation.

Status: Alpaca Paper US and internal B3 simulation confirmed; data in Q8.

## Broker adapter

Proposed contract: get capabilities/instruments, account, open orders, query by client ID, send/cancel order, executions by cursor and positions.
Responses distinguish definitive rejection, temporary limitation, timeout and indeterminate state. `Capabilities` makes explicit the market, mode, order types, time-in-force,
minimum quantity/step, query by ID and idempotent behavior.

Alpaca Paper uses `https://paper-api.alpaca.markets` with paper credentials. The data host is a separate configuration; do not infer that the feed is also on the paper host.
At boot, check mode/account and block live hosts in the MVP. Respect rate limits,
backoff with jitter for reads, and a reconciliation policy for mutations.
Contract tests do not use the personal account in CI.

## Internal B3 adapter

Simulate acceptance, fills, partials, rejections, cancellation and latency with a controlled clock. It has no execution connection to the exchange. Accept only instruments
with tick, lot and session metadata; the final parameters depend on Q7/Q9.
Fill/cost rule versioned, with no implicit unlimited liquidity. Synthetic data
serves fixtures and labeled demonstration, not strategy validation.

## History import — proposed in Q8

CSV will document symbol/market/feed, explicit timezone/offset, timeframe,
open_time, OHLCV, currency and revision. Import with a source/license manifest,
interval, file hash and adjustment policy. Reject conflicting duplicates,
ambiguous times and invalid intervals; report invalid rows without silently importing
only the convenient ones. Limit size and treat file names
as untrusted input. The physical storage form of the original file
on Railway will be defined with retention; the database stores the manifest/hash and candles.

## Interface states and acceptance

Show `Alpaca Paper`, `B3 Simulation — replay` or `B3 Simulation — real time`
according to the clock/source actually implemented. Missing dataset → empty/import action; API disconnected → degraded; stale price → entry block.
No balance or fill may migrate between accounts/modes. Test partial cancellation,
deduplication, missing source, rate limit and crash before/after sending.
