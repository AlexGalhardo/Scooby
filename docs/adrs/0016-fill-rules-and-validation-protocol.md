# ADR-0016: fill rules, resting orders and the validation protocol

State: implemented on 2026-11-04. Updates ADR-0010 and SDD 04. Background:
[docs/trading/backtesting.md](../trading/backtesting.md).

## Context

The backtester filled an order once, on the next bar, canceled the unexecuted balance, and treated a
limit that the open did not reach as canceled. Reference projects either document no fill rule
(Hummingbot paper trading) or model a single price per bar (Freqtrade). A backtest that cannot say
when an order fills cannot be audited, and one test period used to tune and to judge invites
overfitting.

## Decision

Fill rules, applied by `Simulate` and stated in every report:

1. Signals use closed candles only; a signal fills no earlier than the next bar.
2. A marketable order fills at the open plus slippage, never beyond the bar's high (buy) or low (sell).
3. A limit the open does not reach **rests** and fills at the limit only when the bar trades strictly
   through it. Touching the limit does not fill: it gives no queue priority.
4. An order takes at most 1% of a bar's volume (`ParticipationRate`). The rest stays open and keeps
   filling on later bars of the same session; the engine expires it at the session end.
5. When one bar could hit both stop and target, the stop wins; a gap through a stop fills at the open.
6. Costs apply on both sides and slippage is never zero by default.

Validation protocol:

- Partitions are `train` (first 60%), `test` (next 20%), `validation` (last 20%), `all`, and
  `walkforward` (all, plus four consecutive windows that each start cold).
- The API allows the `validation` partition **once per strategy version**; a repeat needs
  `repeat_validation` and is recorded as a decision. A new version is a new candidate.
- Every report carries Sharpe, Sortino (risk-free rate stated, zero), Calmar, SQN, exposure, the
  buy-and-hold baseline, a deterministic block bootstrap of the closed trades, a cost stress run with
  doubled costs, and the deflated Sharpe ratio discounted by the number of versions tried. A run with
  fewer than 30 trades is flagged.

## Alternatives

- Walk the order book for each fill: needs book data this project does not have yet.
- Price-crossing only, without resting orders: simpler, but a limit would never survive a bar.
- One test period: simpler, but the final check would also be the tuning set.

## Consequences

Backtest results change against earlier runs (the old "test" partition was the last 30%). A
golden-file test pins the metrics and a perturbation test proves that no strategy reads future bars.
Market making stays out of scope until order-book data and a validated queue model exist.

## Review trigger

Order-book data becomes available, a live adapter is proposed, or paper results diverge from the
backtest expectation beyond the tolerance recorded for the live gate.
