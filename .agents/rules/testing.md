---
paths:
  - "**/*_test.go"
  - "tests/**"
  - "e2e/**"
---

# Testing rules

- Test first. A ROADMAP.md task is done only when its acceptance commands pass.
- Go: standard library plus `github.com/google/go-cmp/cmp`, table-driven, `t.Parallel()` when safe, `t.Context()`, no testify.
- Unit tests (`go test ./...`) never touch the network or Docker. Integration tests read `TEST_DATABASE_URL` (a database ending in `_test`) and skip with a clear message when it is unset.
- Money, risk and execution code needs a failure-path test next to the happy path: rejected order, stale feed, kill switch, timeout leaving an order `unknown`, concurrent reservations, leader failover.
- Backtests are deterministic: the same dataset hash, configuration and code give the same report. Include a look-ahead test that changes future candles and asserts earlier signals are unchanged.
- Time-dependent code takes a clock; never `time.Sleep` in tests.
- TypeScript: `bun run typecheck` and `bun run lint`. E2E: Playwright in `e2e/` against a local stack; the fake OAuth exists only under the `e2e` build tag and `KindTest`.
- Record the red run before the fix in the report; when a red run is only inferred, say so.
- Run tests on a loaded host with `-p 1`; a rate-limit test that depends on wall time must stretch its window.
