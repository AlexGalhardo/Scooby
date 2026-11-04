# QuantDinger: patterns for vukem-trading

Source: [OpenByteInc/QuantDinger](https://github.com/OpenByteInc/QuantDinger) README. Researched 2026-11-04. Backend Apache-2.0; web and mobile clients are separate source-available repositories. It is the best reference here for the safety model, not for topology.

## What it is

A self-hosted "AI Trading OS": research, Python strategies, backtests, paper and live trading across crypto, stocks and forex, with multi-tenant SaaS pieces. Python 3.12 backend on Flask and Gunicorn; Celery for finite jobs such as backtests; long-lived trading workers; PostgreSQL, Redis and Kafka. About ten process roles, far heavier than this project needs.

## Strategy research

Factor lab with holdout and false-discovery-rate control; strategy search (random, grid, TPE) with walk-forward, blind holdout, probability of backtest overfitting, deflated Sharpe, block-bootstrap Monte Carlo and cost stress. Virtual signal-only accounts use a fixed 0.05% commission and 0.05% slippage and never reach a broker.

## Agent safety model

- An agent gateway and MCP server for coding agents. Agents never receive broker credentials or admin tokens. Agent tokens are hashed, scoped, rate limited and audit logged.
- Agent trading is paper only by default. Live needs a trading-scoped token, an explicit non-paper flag, a server-side environment switch, and operator limits and allowlists.
- An optional AI gate before live entry orders rejects entries only; exits, stops and emergency actions bypass it. It **fails open** when no AI provider is available. This project should do the opposite (fail closed for new entries).
- Deterministic risk and order-budget checks run before the AI gate.
- Strategy ownership uses leases, heartbeats and fencing tokens. Broker credentials are encrypted at rest.
- Documented weakness: a default admin login in the manual Docker install, flagged as not for internet-facing use.

## Recommendations for vukem-trading

1. Paper-only default for any agent surface; live behind a separate server-side flag and scope.
2. Run deterministic risk checks before any LLM involvement; never let an LLM block an exit.
3. Fail closed for new entries when a safety dependency is unavailable.
4. Keep lease, heartbeat and fencing for workers (already implemented here).
5. Add overfitting statistics to backtest reports: deflated Sharpe, probability of backtest overfitting, block-bootstrap Monte Carlo.
6. Do not adopt the Kafka and multi-service topology at this scale.
