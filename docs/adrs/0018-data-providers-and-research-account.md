# ADR-0018: data-provider interface and the research account

State: implemented on 2026-11-04. Extends ADR-0006 and ADR-0008. Background:
[docs/trading/openbb.md](../trading/openbb.md).

## Context

The owner targets B3, US equities, crypto and forex. Data entered the system through two special
cases: a CSV importer for B3 and an Alpaca call for the US. Crypto and forex had no path, price
rounding was fixed at two decimals (wrong for EURUSD), and nothing told the user which market had
live, delayed, imported or no data.

## Decision

- `internal/providers` defines `Provider` (name, markets, required credentials, `Bars`) and a
  `Registry`. A provider declares its credential names; the registry resolves `<PROVIDER>_<NAME>` from
  the environment and returns `ErrMissingCredential` **before** any network call. Adapters:
  `AlpacaIEX` (US) and `Binance` (crypto, public klines, paged, closed candles only).
- Instruments carry a tick and a lot step; prices round to the tick, never to a fixed two decimals.
- Markets `CRYPTO` (continuous UTC sessions, never closes) and `FX` (Monday to Friday in UTC, closing
  out before the weekend) join `B3` and `US`.
- Crypto and FX are **backtest-only**. They run under a `research` account that holds their risk policy,
  never executes, and cannot have a strategy enabled. B3 stays on replay and US on Alpaca Paper.
- A coverage matrix (`trading.CoverageMatrix`) states, per market, the data mode, what can execute and
  the source. It is served in the snapshot and shown in Settings.
- Every provider has a row with its terms in [docs/sources.md](../sources.md); a test enforces it.

## Alternatives

- One account per new market: more screens and engine loops for markets that cannot execute.
- Hide the matrix in the docs: the UI would still imply live data where there is none.

## Consequences

A new provider is a small struct plus a row in `docs/sources.md`. A crypto or forex paper adapter needs a
broker or testnet account to validate and is not part of this decision. The snapshot has three
accounts.

## Review trigger

A crypto or forex paper venue is chosen, or a provider's terms change.
