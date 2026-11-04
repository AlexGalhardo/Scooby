# ADR-0022: change stream, Binance testnet adapter and quote engine

State: implemented on 2026-11-04. Builds on [ADR-0019](0019-workspace-ticket-and-wire-types.md) and
[ADR-0018](0018-data-providers-and-research-account.md).

## Context

Three pieces were left open by earlier decisions: the terminal only polled, there was no crypto paper
venue, and market making had no quoting math. None of them can be validated end to end without something
external (a live feed, a testnet account, an order book).

## Decision

- **Change stream.** `GET /api/stream` (owner session) is a Server-Sent Events endpoint. Every two seconds
  it computes `Store.ChangeSignature`, a fingerprint of accounts, datasets, strategies, backtests,
  artifacts, the workspace and the audit trail, and sends a `changed` event when it moves. The page
  refetches on the event; the ten-second poll stays as the fallback. Because the engine writes its
  heartbeat every 15 seconds, the page also learns about worker activity without waiting.
- **Binance testnet adapter.** `adapters.BinanceTestnet` signs requests with HMAC-SHA256, sends LIMIT GTC
  orders with a client order id, looks orders up by that id so a retry never duplicates one, cancels, and
  reads balances. It **refuses every host except the Binance spot testnet** (loopback is allowed for tests)
  and never echoes a request or a secret in an error. It is not wired into the engine: crypto stays
  backtest-only until a testnet account exists to validate against.
- **Quote engine.** `trading.AvellanedaStoikov` computes a reservation price and a spread from inventory,
  volatility, risk aversion and liquidity, caps inventory per side, and rounds to the tick. It is math only;
  nothing sends the quotes, because market making needs an order book and a queue model (ADR-0016).
- **Live gate and decisions.** `trading.EvaluateLiveGate` and the decision record are described in
  [ADR-0021](0021-live-readiness-gate.md) and [docs/mcp.md](../mcp.md).

## Alternatives

- WebSocket instead of SSE: bidirectional, but the page only needs server-to-client notifications, and SSE
  reconnects by itself.
- Wire the testnet adapter into the engine now: it would execute orders nobody has validated.

## Consequences

The stream issues one cheap query per table every two seconds per open page; for one owner that is
negligible. The adapter and the quote engine are tested against fakes and math, not against a venue.

## Review trigger

A Binance testnet account is available, an order-book source is chosen, or more than a few clients hold
the stream open.
