---
paths:
  - "database/migrations/**"
  - "internal/trading/store*.go"
---

# Database rules

- Expand, migrate, contract: add a column or table in one release, backfill, and drop the old shape later. Never rename in place.
- One migration per logical change, with a `Down` that really reverses it. Concurrent index builds run outside a transaction.
- Every new query path gets its index in the same change; the migration comment names the query it serves.
- Financial state (ledger, fills, reservations, audit) is append-only. Never truncate or rewrite history to save space; archive with a trail instead (ADR-0012).
- Order intents and the outbox are written in the same transaction as the risk decision and reservation.
- `jsonb` columns carry a short comment with the shape they hold and the code that writes it; use `CHECK` constraints for enums instead of free text.
- Tests that need PostgreSQL use a dedicated database whose name ends in `_test` (`TEST_DATABASE_URL`). Never point tests at a shared or production database.
- Never run a migration or a data command against production from an agent session; prepare the exact command for the owner and run it with `--dry-run` first.
