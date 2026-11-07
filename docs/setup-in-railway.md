# Set up in Railway

Four services in one Railway project: **web**, **engine**, **PostgreSQL** and **site** (the documentation published at <https://trading.vukem.dev>). Everything is paper trading.

> Services are connected to the GitHub repository and branch `main`: **a push to `main` deploys**. Use pull requests and wait for green CI before merging.

## Topology

| Service | Source | Public | Notes |
| --- | --- | --- | --- |
| web | Root `Dockerfile`, start `/app/app` | HTTPS domain | Runs migrations in the pre-deploy step; healthcheck `/ready` |
| engine | Root `Dockerfile`, start `/app/engine` | No domain | Paper execution loop; healthcheck `/ready` on `ENGINE_HEALTH_PORT` |
| PostgreSQL | Railway database | Private network | Daily backup; see [backups](#backups) |
| site | `site/Dockerfile`, build context repository root | `trading.vukem.dev` | Astro documentation, static pages served by nginx |

Keep one replica of the engine and never scale it by changing the replica count: leader election and fencing protect against duplicate execution, not against a deliberate fan-out.

## Create the project

1. Install the Railway CLI and log in: `railway login`.
2. Create or link the project: `railway link`.
3. Add PostgreSQL from the Railway dashboard (Database, PostgreSQL).
4. Create the services from the GitHub repository `AlexGalhardo/vukem-trading`, branch `main`.

## Web service

- Builder: Dockerfile, path `Dockerfile`.
- Pre-deploy command: `/app/app migrate`. Start command: `/app/app`. Healthcheck path `/ready`, timeout 120 seconds. Restart on failure, up to 5 times.
- Variables (template: `.env.railway.web.example`; copy it to a `.local` file that is ignored by Git):

  ```text
  APP_ENV=production
  APP_DEBUG=false
  APP_KEY=<generated once, shared with the engine>
  APP_URL=https://<web domain>
  DATABASE_URL=${{Postgres.DATABASE_URL}}
  SESSION_DRIVER=database
  CACHE_STORE=database
  MAIL_MAILER=log
  OWNER_GOOGLE_EMAIL=<your Google account>
  SOCIAL_GOOGLE_CLIENT_ID=<Google OAuth client>
  SOCIAL_GOOGLE_CLIENT_SECRET=<Google OAuth secret>
  ```

- Google callback: `https://<web domain>/auth/google/callback`.
- Optional: `ALPACA_API_KEY` and `ALPACA_API_SECRET` (**paper** credentials only), `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID`.

## Engine service

- Same Dockerfile; start command `/app/engine`; no public domain.
- Variables (template: `.env.railway.engine.example`): the same `APP_KEY`, `DATABASE_URL`, `SESSION_DRIVER`, `CACHE_STORE`, `APP_ENV`, plus Alpaca and Telegram values if used. Railway provides `PORT`; do not copy the local `ENGINE_HEALTH_PORT`.
- Broker credentials live only here and in the web service for the IEX import. Never put them in `VITE_*` variables.

## Site service and DNS

1. Create a service from the same repository. Set `RAILWAY_DOCKERFILE_PATH=site/Dockerfile`. The build context is the repository root because the site reads `docs/` and `CHANGELOG.md`.
2. Add the custom domain `trading.vukem.dev` in the service's Networking settings. Railway shows the DNS target.
3. Create the record at your DNS provider exactly as Railway shows it (a `CNAME` for `trading` pointing at the Railway target, plus the verification `TXT` record if requested). Check it before applying: a wrong record can take down another service on the same zone.
4. Wait for the certificate and open <https://trading.vukem.dev>. The root path redirects to the introduction page.

## Deploy, rollback and logs

- Deploy: merge a pull request into `main` after CI is green. Manual redeploy: `railway redeploy`.
- Status and logs: `railway deployment list --json`, `railway logs`.
- Rollback: redeploy the previous successful deployment from the dashboard, or revert the commit.
- Smoke test after each deploy: `/` returns 302 to login, `/login` 200, `/ready` 200, `/api/snapshot` without a session 401.

## Backups

Enable the daily PostgreSQL backup schedule in the dashboard and confirm the provider's retention (six days at the time of writing; complement it if you need seven). Run a restore into a scratch database before you rely on it. `scripts/backup.sh` creates an off-platform dump.

## Costs and region

Railway bills by usage. Confirm region, plan, CPU and memory, and a monthly budget alert before enabling the engine continuously. The region is inherited from the project; do not assume TimescaleDB is available.

## Checklist

- [ ] PostgreSQL attached and the `Postgres` reference name matches `DATABASE_URL`.
- [ ] Same `APP_KEY` in web and engine; generated with `vukem key:generate`, never committed.
- [ ] Google OAuth callback matches the web domain.
- [ ] Engine has no public domain and one replica.
- [ ] Healthchecks pass; smoke test done.
- [ ] Backup schedule enabled and a restore tested.
- [ ] `trading.vukem.dev` resolves and serves the documentation.
