# SDD-05 — Railway delivery and operation

> Current implementation and explicit deviations are in the [ROADMAP](../../ROADMAP.md)
> and [ADRs 0011–0013](../adrs/README.md). Local technical acceptance is distinct
> from real OAuth/broker, Railway publication and prolonged financial observation.

Status: Railway confirmed; design and targets proposed.

## Publication contract

Versioned image, compatible migrations, secrets per service and separation of
web/worker. The engine must remain active without HTTP traffic. Database shared
on a private network. No change to paper/live endpoints by automatic deploy decision.
The worker enters by reconciling, not in immediate execution mode.

Future artifacts: Railway configuration per service, build of the two binaries,
readiness, migration command with lock, metrics, alerts and runbooks. There is no
engine service nor Railway configuration implemented in the current starter.

## Acceptance scenarios

- Deploy with a failing migration does not promote the new web/engine version.
- Overlapping workers do not allow duplication of intent/sending by the new leader.
- SIGTERM preserves the checkpoint and does not lose fills received during shutdown.
- Restart keeps the kill switch, daily loss and unknown orders.
- Absence of a browser session does not interfere with the heartbeat.
- Readiness fails when essential dependencies prevent serving/operating; liveness
  is distinct and does not disclose private data.
- Restore tested on an account without execution; reconcile before resuming.
- Alerts for stale feed, stopped worker and ambiguous order work without the dashboard.

## Operational pending items

RPO, RTO, SLO, region, budget, retention, Telegram alert or other channel,
load limit and incident owner depend on the operational round.
Details and variables in `../railway.md`; do not copy secrets into SDDs.
