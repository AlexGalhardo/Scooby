# Web terminal and agent boundaries

Synthesis of [eTape](etape.md), [OpenTerminal](openterminal.md), [OpenStock](openstock.md), [OpenBB](openbb.md) and [QuantDinger](quantdinger.md).

## Terminal features by priority

1. **Safety surfaces first:** master arm switch, kill switch, live risk panel (exposure, daily loss, open orders), account state banner (blocked, reconciling, ready).
2. **Core trading views:** chart with fill markers, positions, orders, audit log, backtest lab with report export (largely present).
3. **Workspace:** dockable widgets with saved layout, command palette, numbered shortcuts, linked symbol groups.
4. **Market views:** watchlist (capped size), coverage matrix per market, screener, session-aware gainers and losers.
5. **Fast entry (paper):** hotkey order templates sized by percentage of buying power or position, chart risk entry that sizes from a loss budget.
6. **Alerts:** threshold alerts with a checker interval and a delivery channel (Telegram exists).

## Frontend engineering rules

- High-frequency data (tape, ladder) is drawn on canvas outside React state; push updates over SSE or WebSocket, with polling as fallback.
- Generate TypeScript wire types from Go structs and fail CI on drift.
- Never place credentials in props, bundles or `VITE_*` variables.
- Every number shows its currency and its data freshness.

## Agent and AI boundaries

- An LLM never signs, sends or edits orders. It may draft a validated strategy configuration for human review.
- Any agent surface starts read-only and paper-only, with hashed, scoped, rate-limited, audit-logged tokens. Live trading, if it ever exists, needs a separate server-side flag.
- Deterministic risk runs before any model; a model can only reject entries and never blocks exits; if the model is unavailable, new entries fail closed.
- MCP exposure, if added, is read-only data tools with a bearer token and a tool allowlist.
- Never claim an AI can detect tops or bottoms, or promise returns.
