# ADR-0005 — Railway hosting

Date: 2026-11-03. State: **Accepted as destination; proposed topology**.

## Context

The user chose Railway for deploy; execution requires a continuous service.

## Decision

Host on Railway. Propose separate web and engine services, private PostgreSQL, migrations in pre-deploy and an engine without hibernation.

## Alternatives

VPS, Kubernetes or execution in ephemeral functions. These alternatives are not the requested destination; they may be revisited if measured limitations require it.

## Consequences

Configure health, secrets, resources, concurrent deploy and backup. Railway does not by itself guarantee executor exclusivity or reconciliation.

## Review trigger

Limitation of region, connectivity, cost, availability or specific broker requirements.
