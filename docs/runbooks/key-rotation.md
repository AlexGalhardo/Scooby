# Runbook: rotate keys and credentials

Rotate on a schedule you choose and immediately after any suspected exposure. Never paste a secret into
chat, an issue or a commit; set it through the provider's variables.

## Application key (`APP_KEY`)

The web and the engine must always share the same key.

1. Generate the new key with `vukem key:generate` on a machine you control.
2. Set `APP_PREVIOUS_KEYS` to the **old** key in both services, and `APP_KEY` to the **new** key.
3. Deploy web and engine together. Data encrypted with the old key still decrypts.
4. After a day, remove `APP_PREVIOUS_KEYS`.

## Google OAuth

Create a new client secret in Google Cloud, set `SOCIAL_GOOGLE_CLIENT_SECRET` on the web service, deploy,
sign in once to confirm, then delete the old secret. The redirect URI does not change.

## Alpaca Paper

Generate a new key pair in the Alpaca dashboard (paper account), set `ALPACA_API_KEY` and
`ALPACA_API_SECRET` on the web and the engine, redeploy, and confirm the safety strip shows a fresh
reconciliation. Revoke the old pair. Never configure live credentials: live hosts are rejected.

## Telegram

Create a new bot token with BotFather, set `TELEGRAM_BOT_TOKEN` (and check `TELEGRAM_CHAT_ID`), redeploy
the engine, then revoke the old token. The uptime workflow reads the same names as repository secrets.

## Database password

Railway manages the PostgreSQL credentials and exposes them through `DATABASE_URL`. To rotate, reset the
credentials in the database service and redeploy the services that reference it.

## After any rotation

Open `GET /api/metrics`: heartbeat and reconciliation ages must be fresh. Check the audit page for
unexpected `risk.halt` entries.
