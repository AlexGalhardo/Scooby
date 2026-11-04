# ADR-0021: live-readiness gate

State: **proposed**. It needs the owner's sign-off (see the last section). No live adapter exists and none is
authorized by this ADR. Background: [docs/trading/risk-and-execution.md](../trading/risk-and-execution.md),
[ADR-0007](0007-paper-first.md).

## Context

The owner chose "paper now, live later behind gates". Reference projects that also trade live (Hummingbot,
QuantDinger) show how easily a convenience flag becomes a way to lose money: a setting an assistant or a UI
can flip, credentials in the web tier, or generous first limits.

## Decision

Live trading may be considered only when **all** of these hold, each with a named proof:

| Criterion | Proof |
| --- | --- |
| Paper results over an agreed period match the backtest within a stated tolerance | A paper observation of at least 30 trading days per strategy, compared with `trading.WithinTolerance` at a 25% relative tolerance (one-point floor), recorded in the audit trail |
| The risk gate, kill switch and reconciliation have failure-path tests in CI | Default deny, mandatory stop, daily-loss kill switch, protections, concurrent reservations, rollback, timeouts and leader failover are tested today |
| Live is enabled by a server-side switch that no owner or assistant setting can change | `trading.EvaluateLiveGate` reads only the process environment: `LIVE_TRADING_ENABLED=true` and the exact acknowledgement sentence. The workspace, the risk policy, the MCP endpoint and the UI are not inputs |
| Credentials exist only in the worker | The gate returns off for any role but `worker`, so the web service can never enable live trading |
| First-period limits are deliberately low | The gate requires `LIVE_MAX_ORDER_VALUE` of at most 100 and `LIVE_MAX_DAILY_LOSS_BPS` of at most 25; raising those ceilings is a code change |
| The owner signs off in this ADR | The sign-off block below |

A live adapter must call `EvaluateLiveGate` before it starts and refuse to run when `Enabled` is false.
Live hosts for Alpaca are already refused by the paper adapter; a live adapter is a separate package.

## Alternatives

- A runtime setting in the UI: convenient, and exactly what an assistant or a mistaken click could change.
- Credentials in the web tier: the web already holds Alpaca paper keys to import data; live keys must not follow.

## Consequences

Until the observation period exists and the owner signs, the gate stays closed. The checklist in the
ROADMAP (Phase 17) tracks the criteria; the ones needing elapsed time or a signature remain open.

## Owner sign-off

Not given. To sign: set the state above to "accepted", record the date, the observation period and the
tolerance result, and commit the change yourself.
