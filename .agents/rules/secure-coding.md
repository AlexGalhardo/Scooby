---
paths:
  - "**/*.go"
  - "ROADMAP.md"
---

# Security rules

- Secure by default; unsafe behavior is an explicit opt-out with an "Insecure" name, never a missing opt-in.
- Fail closed: if a check cannot run (store down, policy missing, key missing, feed stale), deny. New entries are rejected; reducing-risk exits follow their own policy.
- Misconfiguration that weakens security is a boot error, not a warning.
- Broker, Google and Telegram credentials exist only in the engine worker or the server; never in props, bundles, logs, fixtures or Git history.
- Never log secrets, tokens, `Authorization` or `Cookie` headers.
- SQL: bound parameters only; never concatenate input into SQL. HTML through `html/template`.
- Randomness from `crypto/rand`; compare secrets with `subtle.ConstantTimeCompare` or `hmac.Equal`.
- Client-facing errors never include stack traces, SQL, paths or configuration.
- File access through `os.Root`; outbound requests to user-supplied URLs through the vukem safe HTTP client.
- CSV and dataset imports validate size, encoding, timestamps and decimals before any write, and are atomic.
- Every security behavior names its proof: `TestSecurity_<Axx>_<Behaviour>`.
- Load the `golang-security` skill for changes in accounts, uploads, webhooks, integrations and API keys.
- Any agent or MCP surface is read-only and paper-only by default, uses scoped, hashed, rate-limited tokens, and never receives broker credentials. An LLM never signs or sends orders.
- Production data commands run with `--dry-run` first and only with the owner's yes.
