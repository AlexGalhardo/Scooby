# Runbook: back up and restore PostgreSQL

All state lives in PostgreSQL: accounts, the ledger, datasets, strategies, backtests, audit events and the
saved workspace. Back it up, and prove the backup restores.

## Backups

- Railway runs a daily PostgreSQL backup schedule; confirm its retention in the dashboard (six days at the
  time of writing).
- For an off-platform copy run `scripts/backup.sh` against the database URL; it writes a custom-format
  `pg_dump` file. Keep at least seven dumps, outside Railway.

## Restore into a scratch database first

Never restore over production blindly.

```sh
createdb restored
pg_restore --no-owner -d restored backup.dump
```

Check that every table has the expected row count, and that the newest `audit_events` row is the one you
expect.

## Restore production

1. Halt every account (Risk control, or the kill switch) so no new decisions are written.
2. Stop the engine service so it cannot write during the restore.
3. Restore into an empty database, point `DATABASE_URL` at it, and start the web service. It applies any
   missing migrations before serving.
4. Start the engine. It increments the leader epoch, marks in-flight sends as `unknown` and reconciles
   before it allows entries.
5. Compare the broker's paper account with the Orders and Positions pages, then resume (see
   [incident response](incident.md#4-reconcile-and-resume)).

## Rehearsal

`scripts/restore-rehearsal.sh` rehearses the whole cycle on a throwaway container: migrate, add rows,
dump, restore, compare every table. Run it after any schema change and before relying on a backup policy.
The latest recorded run is in [restore-rehearsal-2026-11-04.md](restore-rehearsal-2026-11-04.md).
