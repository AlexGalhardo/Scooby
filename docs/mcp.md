# Read-only MCP endpoint

The terminal can expose its market data, reports and operational metrics to an assistant through the
Model Context Protocol, over a single HTTP endpoint (JSON-RPC 2.0). It can read. It cannot trade.

## Turn it on

Set a long random `MCP_TOKEN` on the web service. Without it the endpoint answers `404`. Generate the
token on a machine you control and never paste it into chat or an issue.

```sh
openssl rand -base64 32
```

Every request must carry `Authorization: Bearer <token>`. A missing or wrong token answers `401`; the
comparison is constant-time.

## Endpoint

`POST /mcp` on the web service. Implemented methods: `initialize`, `ping`, `tools/list`, `tools/call`
and the `notifications/initialized` notification. Request bodies are limited to 64 KiB.

## Tools

The list is fixed in code (`app/controllers/terminal/mcp.go`) and a test pins it.

| Tool | Returns |
| --- | --- |
| `list_markets` | Coverage per market: data mode, what can execute, source, notes |
| `list_instruments` | Instruments with market, currency, lot and tick |
| `list_strategy_kinds` | Strategy kinds and the declared range of each parameter |
| `list_datasets` | Dataset metadata: symbol, source, license, hash, bars, period (no candles) |
| `get_backtest_report` | Summary of a finished backtest: metrics, robustness, walk-forward windows (no curve or fills) |
| `get_ops_metrics` | Heartbeat, reconciliation, rejection rate and slippage of every account |

## What an assistant cannot do

There is no tool that places, changes or cancels an order, changes a risk policy, starts or stops an
account, or enables a strategy. A test fails if the tool list ever contains such a name, and the endpoint
carries no session, so it also cannot reach the owner-only routes.

## Strategy drafts

An assistant can help write a configuration, but it enters the system through the owner's session: paste it
to `POST /api/strategies/draft`. It is validated like any other strategy and stored as a **draft**. A draft
cannot be enabled until the owner approves it explicitly (the "Approve and enable" button, or
`approve_draft: true`), and the approval is audited as `strategy.draft.approved`.
