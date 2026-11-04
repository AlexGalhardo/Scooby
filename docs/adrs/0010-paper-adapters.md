# ADR-0010 — Alpaca Paper and internal B3 simulator

Date: 2026-11-03. State: **Accepted by answers Q4 and Q5; implementation pending**.

## Context

The MVP covers Brazil and the US without real money. There is no proven B3 order integration or
selected Brazilian broker.

## Decision

Use Alpaca Paper as the first US integration. B3 starts with backtest and an
internal simulator; an external paper integration only after validating the broker's API/contract. Data adapter separate from the execution one. The IEX feed and B3 CSV
import are Q8 proposals, not automatically accepted as part of this ADR.

## Alternatives

IBKR, requiring a B3 broker immediately, or supporting only the US. The first two
were not chosen; the last conflicts with the two confirmed markets.

## Consequences

The dashboard distinguishes broker paper orders from internal simulated fills. These are
separate accounts/ledgers. Without a data provider, the internal simulator does not receive
real quotes by magic: history and real time need a defined source.
Fill rules, costs and reviewed data must be included in each report.

## Review trigger

Verified B3 API, limitation of the Alpaca plan or request for another broker. Implement a
new adapter with contracts; do not switch the paper URL to live.
