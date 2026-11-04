# SDD-02 — market data and strategies

> Current implementation and explicit deviations are in the [ROADMAP](../../ROADMAP.md)
> and [ADRs 0011–0013](../adrs/README.md). Local technical acceptance is distinct
> from real OAuth/broker, Railway publication and prolonged financial observation.

Status: Brazil + US confirmed; contracts proposed. Brokers, feeds, universe and timeframe pending.

## Input and contract

Market event: `event_id`, `schema_version`, `source`, `instrument_id`,
`event_time`, `received_at`, sequence when available and decimal payload.
Candle: timeframe, open/close time, OHLCV, closed, feed and revision. Reject
NaN/Inf, invalid prices, incoherent OHLC, negative volume and an instrument without
metadata. Do not replace missing data with a zero quote or an invented price.

The adapter declares capabilities: history, streaming, delay, calendar, limits and
adjustments. Persist origin and revision. Deduplicate, detect gaps and reorder within a
documented tolerance; beyond it, repair the history and do not emit retroactive signals
as if they were current opportunities. Bootstrap requires a complete warmup.

## Pure and versioned strategies

Logical contract: `Evaluate(context, closed_candles, params, portfolio_snapshot)`
returns `Signal` or `NoSignal`, without external I/O. `Signal` has origin, version,
instant, instrument, direction, reason and validity; it is not yet an approved order.
Window, warmup and validation belong to the version. The same input produces the same output.

Initial proposal: moving average crossover; mean reversion via Z-score; breakout
by band and volatility. Define SMA vs EMA, windows, confirmation, exit,
frequency and filters after market analysis. Zero deviation in Z-score results in no-signal.
No strategy uses a future or still-open candle without a specific contract.

Configuration is declarative JSON/DSL with `schema_version` and an allowlist of indicators,
operators and intervals. Reject unknown fields and invalid limits. No `eval`, user-submitted
scripts or execution of LLM text. Changes
create immutable versions; activation records the actor and the risk policy version.

## Acceptance and failures

Test warmup, repetition, gaps, DST, out-of-order candle, revision and reprocessing.
One bar does not produce two intents for the same execution/versioning. Stale feed
suspends new exposures. The strategy does not call the broker directly.
Obtaining candles does not grant a license for redistribution to users of a future SaaS.
