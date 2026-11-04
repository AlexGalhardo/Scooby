# Restore rehearsal: 2026-11-04

Command: `bash scripts/restore-rehearsal.sh` on the commit that added the script. Image
`postgres:18.6-alpine3.24` in a throwaway container; the migrations came from `go run ./cmd/app migrate`.

Rows added to the source before the dump: one `preferences` row and one `audit_events` row.

Result, as printed by the script:

```text
tables restored: 17
access_tokens 0
artifacts 0
audit_events 1
backtests 0
cache 0
datasets 0
migrations 9
owner_identity 1
password_resets 0
posts 0
preferences 1
remember_tokens 0
sessions 0
social_accounts 0
strategies 0
trading_accounts 0
users 0
PASS: every table has the same row count after the restore
```

What this proves: a custom-format dump of a migrated database restores completely into an empty database
with the same schema and rows. What it does not prove: restore time on production-sized data, or the
provider's own backup. Repeat the check against a real Railway backup after the first scheduled backup
runs (ROADMAP item in Phase 5).
