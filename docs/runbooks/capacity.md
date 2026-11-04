# Capacity, latency, RPO and RTO

Measured on 2026-11-04 on a developer machine (AMD Ryzen 7 5700X3D, 16 threads, Windows, Go 1.27.3), with
`go test ./internal/trading -run '^$' -bench . -benchtime 200ms -benchmem`. These are single-process CPU
numbers, not production numbers; they show where the limits are, not what Railway will deliver.

## Decision latency

| Operation | Time | Allocations |
| --- | --- | --- |
| One risk decision (`Reserve`: sizing, limits, reservation) | about 12 microseconds | 343 |
| Signal, breakout | about 0.7 microseconds | 10 |
| Signal, moving-average cross | about 14 microseconds | 510 |
| Signal, mean reversion | about 12 microseconds | 410 |
| Signal, opening range | about 18 microseconds | 81 |
| Signal, DCA grid | about 8 microseconds | 220 |
| Signal, VWAP reversion | about 130 microseconds | 2,821 |
| Signal, ADX trend | about 290 microseconds | 6,961 |

Signals run over a 140-bar window. The engine ticks every 15 seconds, so computation is never the
bottleneck: a tick is dominated by the database commit and, for the US account, the broker round trips.

## Backtest capacity

A full backtest includes the robustness checks (a second run with doubled costs and the bootstrap).

| Sample | Bars | Time | Memory churn |
| --- | --- | --- | --- |
| 5 sessions | 350 | 0.18 seconds | 73 MB |
| 20 sessions | 1,400 | 0.71 seconds | 256 MB |
| 60 sessions | 4,200 | 1.87 seconds | 761 MB |

Time grows roughly linearly with bars. Memory churn is high: each bar copies parts of the account state.
The importer accepts up to 100,000 candles, which extrapolates to about 45 seconds and about 18 GB of
allocation churn for one job. Jobs run one at a time under a timeout, so a very large sample can time
out or pressure memory on a small Railway plan. Reducing the per-bar copies is a worthwhile optimization
before importing samples of that size; until then, prefer samples under about 20,000 bars.

## Recovery

- **RPO** (data you can lose): the Railway backup schedule is daily, so the RPO is up to 24 hours unless
  you also take `scripts/backup.sh` dumps more often. The ledger is append-only, so a restore never
  rewrites history; it only loses what happened after the dump.
- **RTO** (time to recover): measured on the rehearsal database, `pg_dump` took 433 ms and `pg_restore`
  516 ms, including the container overhead, for 17 tables with a handful of rows. Real restore time grows
  with data size and is not measured here. End to end, plan for: detection (the uptime check runs every 30
  minutes), the restore, a redeploy of the services (a few minutes), and the reconciliation before the
  account is resumed (see [incident response](incident.md)).

Repeat these measurements on the Railway plan in use and after any large change to the engine. The
rehearsal is in [restore.md](restore.md).
