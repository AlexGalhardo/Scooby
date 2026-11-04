# Set up on a VPS with Docker

Run vukem-trading on a single Linux server with Docker Compose and automatic HTTPS. The same image is used by Railway, Fly and plain Docker.

## Requirements

- A Linux VPS with Docker and Compose v2, at least 1 vCPU and 1 GB of RAM.
- A DNS record pointing your domain (for example `trading.example.com`) at the server.
- A Google Web OAuth client with the redirect URI `https://trading.example.com/auth/google/callback`.

## Layout

```text
/opt/vukem-trading/
  compose.yaml
  Caddyfile
  .env
```

## compose.yaml

Pin image tags. Replace `OWNER/REPO` with `alexgalhardo/vukem-trading` and `X.Y.Z` with a released version.

```yaml
name: vukem-trading

services:
  caddy:
    image: caddy:2.11.8-alpine
    restart: unless-stopped
    ports: ["80:80", "443:443"]
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy-data:/data
    depends_on: [app]

  app:
    image: ghcr.io/OWNER/REPO:X.Y.Z
    restart: unless-stopped
    env_file: [.env]
    entrypoint: ["/bin/sh", "-c", "/app/app migrate && exec /app/app"]
    expose: ["3000"]
    depends_on:
      postgres:
        condition: service_healthy

  engine:
    image: ghcr.io/OWNER/REPO:X.Y.Z
    restart: unless-stopped
    env_file: [.env]
    environment:
      ENGINE_HEALTH_PORT: "3001"
    entrypoint: ["/app/engine"]
    depends_on:
      app:
        condition: service_started

  postgres:
    image: postgres:18.6-alpine3.24
    restart: unless-stopped
    environment:
      POSTGRES_USER: trading
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
      POSTGRES_DB: trading
    secrets: [postgres_password]
    volumes: ["postgres:/var/lib/postgresql"]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U trading -d trading"]
      interval: 5s
      retries: 20

secrets:
  postgres_password:
    file: ./postgres_password.txt

volumes:
  caddy-data:
  postgres:
```

PostgreSQL is not published to the host. Create `postgres_password.txt` with a long random password and `chmod 600` it.

## Caddyfile

```text
trading.example.com {
	encode zstd gzip
	reverse_proxy app:3000
}
```

## .env

```text
APP_ENV=production
APP_DEBUG=false
APP_KEY=<output of: vukem key:generate>
APP_URL=https://trading.example.com
DATABASE_URL=postgres://trading:<password>@postgres:5432/trading?sslmode=disable
SESSION_DRIVER=database
CACHE_STORE=database
MAIL_MAILER=log
OWNER_GOOGLE_EMAIL=you@example.com
SOCIAL_GOOGLE_CLIENT_ID=...
SOCIAL_GOOGLE_CLIENT_SECRET=...
ALPACA_API_KEY=
ALPACA_API_SECRET=
```

The web and the engine **must share the same `APP_KEY`**. Keep `.env` at mode 600 and out of Git.

## Start and verify

```sh
docker compose pull
docker compose up -d
docker compose ps
curl -fsS https://trading.example.com/up
```

## Backups

Dump PostgreSQL daily and copy the dump off the server:

```sh
docker compose exec -T postgres pg_dump -U trading -Fc trading > "backup-$(date +%F).dump"
```

Test a restore on a scratch database before you rely on it, and keep at least seven days of dumps.

## Upgrades and rollback

1. Read the release notes of the new version.
2. Back up, change the image tag in `compose.yaml`, then `docker compose pull && docker compose up -d`.
3. Roll back by restoring the previous tag. Migrations that dropped data need the backup.

## Hardening checklist

- Firewall: allow 22 (restricted to your IP), 80 and 443 only.
- SSH keys only; no password login.
- Unattended security updates enabled.
- The engine has no public port; keep it that way.
- Use `docker compose logs` and an external uptime check on `/up`.
