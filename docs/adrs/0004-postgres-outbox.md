# ADR-0004 — financial persistence and outbox

Date: 2026-11-03. State: **Proposed; PostgreSQL confirmed**.

## Context

Orders and reservations must be durable and consistent; the external network does not participate in the database transaction.

## Decision

Propose PostgreSQL as the record of intents, fills, ledger, reservations and the transactional outbox. NUMERIC/decimal for values; uniqueness of signals and execution IDs.

## Alternatives

Redis as the definitive database; messaging without outbox; floats in accounting. They do not offer the invariants required by the proposed design.

## Consequences

Requires per-account locks, indexes, idempotent workers and outbox cleanup/monitoring. Does not deliver exactly-once at the broker; timeout requires reconciliation.

## Review trigger

Measured contention or volume requiring another transport, while preserving the record and invariants.
