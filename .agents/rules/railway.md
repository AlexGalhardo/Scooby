---
paths:
  - "railway*.toml"
  - "Dockerfile"
  - "docs/setup-in-railway.md"
  - "docs/railway.md"
---

# Railway rules

Read [.agents/skills/use-railway/SKILL.md](../skills/use-railway/SKILL.md) first. Never deploy or change DNS unless the owner asks.

- A root `railway.toml` applies to every service built from the repository: keep it to shared settings and choose each service's image with the `RAILWAY_DOCKERFILE_PATH` variable. Per-service config-as-code files are deprecated; set service settings (start command, health check, volume) in the dashboard or API.
- `app.trading.vukem.dev` is the `web` service; the paper engine is a separate service with its own start command.
- Custom domain: add it on the Railway service, then a CNAME and the `_railway-verify` TXT record. `vukem.dev` DNS is on Vercel (`vercel dns add`).
