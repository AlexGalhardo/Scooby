# Scooby

Projeto web com [Vukem](https://vukem.dev/docs/introduction) (Go), Inertia.js,
React, TypeScript, Tailwind CSS e PostgreSQL. Criado com o starter oficial
`vukem new Scooby --module github.com/AlexGalhardo/Scooby --inertia react --pm npm --docker --ci`.

## Desenvolvimento

Requisitos: Go 1.27.3+, Node.js 24+ com npm e Docker Compose.
Todas as ferramentas podem rodar no ambiente de nuvem do Codex.

```sh
cp .env.example .env
go install github.com/AlexGalhardo/vukem/cmd/vukem@v0.37.0
vukem key:generate
npm ci
docker compose up -d postgres
vukem migrate
vukem serve
```

Abra `http://localhost:3000`. O comando `vukem serve` inicia Go e Vite com
recarga automática. Sem a CLI, execute `npm run dev` e `go run ./cmd/app`
em terminais separados.

A página inicial e o CRUD de posts (`/posts`) demonstram a integração Inertia.
O PostgreSQL é o banco padrão: o driver está registrado em `bootstrap/app.go`
e `DATABASE_URL` está em `.env`. SQLite em memória é usado apenas pelos testes
isolados do starter.

O `.env` local já recebe uma chave aleatória na criação e nunca entra no Git.
As credenciais `scooby:scooby` são exclusivas de desenvolvimento. Em produção,
configure uma chave própria, `APP_ENV=production`, `APP_DEBUG=false`, `APP_URL`
e uma `DATABASE_URL` segura com TLS conforme seu provedor. O CRUD de exemplo
é público; adicione autenticação e autorização antes de armazenar dados privados.

## Stack completa em Docker

Após criar `.env`, gerar `APP_KEY` e instalar as dependências:

```sh
docker compose up --build
```

O Compose inicia a aplicação, aplica migrations, sobe PostgreSQL e disponibiliza
o Mailpit em `http://localhost:8025`. O banco persiste em volume Docker.
Para parar sem apagar os dados, use `docker compose down`.

## Verificação

```sh
npm run typecheck
npm run build
go test -race ./...
go vet ./...
go tool govulncheck ./...
```

Para testar migrations e o CRUD Inertia no PostgreSQL real, use um banco
dedicado de testes (nunca produção):

```sh
docker compose up -d postgres
docker compose exec postgres createdb -U scooby scooby_test
TEST_DATABASE_URL='postgres://scooby:scooby@localhost:5432/scooby_test?sslmode=disable' go test -v ./tests -run TestPostgresPosts
```

O GitHub Actions verifica Go, tipos TypeScript, build frontend, integração com
PostgreSQL, vulnerabilidades, segredos e build da imagem. `package-lock.json`
e `go.sum` fixam as dependências para instalações reproduzíveis.

## Build

```sh
vukem build  # assets e binário em bin/app
```

O Dockerfile também produz uma imagem com assets embutidos e usuário sem
privilégios. Para publicar, configure as variáveis de produção e execute
`/app/app migrate` antes de iniciar `/app/app`.

## Documentation

Every section of <https://vukem.dev/docs>, in the order a new project needs
them:

| Section | What it covers |
| --- | --- |
| [Getting started](https://vukem.dev/docs/introduction) | installation, the starters, `.env` and configuration, the directory structure |
| [Architecture](https://vukem.dev/docs/architecture/lifecycle) | the application lifecycle, from `cmd/app` to a response |
| [The basics](https://vukem.dev/docs/basics/routing) | routing, middleware, controllers, requests, validation, errors, sessions, CSRF, views |
| [APIs](https://vukem.dev/docs/rest-apis/overview) | JSON APIs, pagination, typed handlers, OpenAPI, client generation, idempotency |
| [Database](https://vukem.dev/docs/database/overview) | the ORM, relations, migrations, factories and seeders, PostgreSQL, Redis, MongoDB |
| [Security](https://vukem.dev/docs/security/overview) | authentication, API tokens, two-factor, authorization, rate limiting, headers, encryption |
| [Digging deeper](https://vukem.dev/docs/digging-deeper/mail) | mail, cache, storage, uploads, events, queue, scheduler, broadcasting, billing, admin |
| [Frontend](https://vukem.dev/docs/frontend/htmx) | HTMX, Vite, Inertia |
| [Testing](https://vukem.dev/docs/testing/overview) | the fluent HTTP client, the test database, factories |
| [CLI](https://vukem.dev/docs/cli/overview) | every `vukem` command and flag |
| [Deployment](https://vukem.dev/docs/deployment/single-binary) | the single binary, Docker and Railway |

## Project standards

| File | What it does |
| --- | --- |
| `.editorconfig` | tabs of width 4, spaces where the format needs them (YAML, Markdown, JavaScript) |
| `.gitattributes` | LF in the repository and in every working copy, Windows included |
| `.betterleaks.toml` | secret scanning rules (`APP_KEY`, `vk_` API tokens, Stripe keys) on top of Betterleaks' defaults |
| `.github/workflows/ci.yml` | format, vet, test, `govulncheck`, secret scanning and the image build on every push |
| `.github/dependabot.yml` | weekly dependency and action updates, with Conventional Commits titles |
| `compose.yaml` | the local stack: the app, PostgreSQL and a mail catcher (`docker compose up --build`) |
