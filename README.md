# Scooby

**Comece pelo [ROADMAP.md](ROADMAP.md)**: onboarding, estado real, testes,
limitações e próxima ação para sessões e agentes.

Terminal pessoal de trading quantitativo em Go/Vukem, Inertia/React/TypeScript,
Tailwind/shadcn-ui e PostgreSQL. Exclusivamente backtesting e paper trading:
Alpaca Paper nos EUA e replay interno B3. Sem capital real.

## Implementado

Login exclusivo Google com conta única e subject persistente; dashboard privado
com sidebar responsiva, gráficos, dados, estratégias versionadas, laboratório de
backtests, ordens, posições, risco, auditoria e status de integrações.

Três estratégias (médias, reversão e rompimento), cálculo decimal, datasets
com hash/procedência, jobs canceláveis e relatórios exportáveis. Worker Go
independente da web com liderança/fencing, reservas e intenções transacionais,
reconciliação de ordens ambíguas/parciais e bloqueio de risco durável.

O [checklist Railway](docs/deploy-checklist.md) contém todas as variáveis e o
procedimento de publicação. Credenciais Google/Alpaca/Railway ainda são externas.

## Desenvolvimento

Requisitos: Go 1.27.3+, Node.js 24+, npm, Docker Compose. Em clone novo, copiar
`.env.example` para `.env` **somente se não existir**. Nunca sobrescrever chave ou
configuração existente. Instalar CLI e gerar chave na primeira preparação:

```sh
go install github.com/AlexGalhardo/vukem/cmd/vukem@v0.37.0
vukem key:generate
npm ci
docker compose up -d postgres
vukem migrate
vukem serve
```

Configurar `OWNER_GOOGLE_EMAIL` e OAuth Google Web em `.env`; callback local:
`http://localhost:3000/auth/google/callback`. Google verifica a identidade e o
servidor admite apenas a conta configurada. Sem OAuth o login informa a pendência.

Em outro terminal, executar o worker:

```sh
ENGINE_HEALTH_PORT=3001 go run ./cmd/engine
```

Ou `go run ./cmd/app trading:work`. Para executar tudo em containers:
`docker compose up --build`. A web aplica migrations antes da engine iniciar.
A engine não precisa de domínio público. Contas começam bloqueadas e estratégias
desabilitadas; ative explicitamente depois de revisar os limites.

## Dados e simulações

Watchlist inicial: PETR4, VALE3, ITUB4, AAPL, NVDA, TSLA. CSV B3 2026 com
`time,open,high,low,close,volume`, timestamps RFC3339 e offset, candles encerrados
de cinco minutos e origem/licença declaradas. Importações são atômicas e rejeitam
duplicatas, ordem temporal inválida, decimais inválidos e fora de sessão.
Calendário B3 2026 inclui feriados/abertura especial e mudança de grade em novembro;
anos não revisados falham fechados. EUA usa calendário e feed IEX Alpaca.

Amostra de formato em [examples/b3-format.csv](examples/b3-format.csv),
**sintética e insuficiente para validar estratégia**. Um único replay B3 ativo por
conta preserva o relógio histórico. EUA permite uma configuração ativa por ativo.
Não somar BRL/USD nem apresentar replay como execução em corretora.

Backtests usam próximo evento elegível, custos/slippage configurados, participação
limitada por volume e stop conservador em ambiguidade OHLC. Treino/teste: 70/30;
benchmark bruto e limitações identificados. Não há promessa de rentabilidade.

## Verificação

```sh
npm run typecheck
npm run build
go test -race ./...
go vet ./...
go tool govulncheck ./...
```

Para integração PostgreSQL, criar um banco dedicado cujo nome termine em `_test`
e configurar `TEST_DATABASE_URL`. Nunca usar produção nesses testes. Casos
incluem concorrência de reservas, rollback e failover de liderança.

```sh
npx playwright install chromium
npm run test:e2e
```

E2E usa servidor loopback com fake OAuth **somente** em `KindTest` e build tag
`e2e`; não é incluído na imagem de produção. Exercita login, importação, criação
de estratégia, job/relatório de backtest, navegação, mobile e logout.
No ambiente Codex com Chromium instalado, usar `CHROMIUM_PATH=/usr/bin/chromium`.
Não rodar `npm run build` ao mesmo tempo que Go compila os assets embutidos.

## Build e operação

`scripts/build.sh` gera `bin/app` e `bin/engine`, com versão Git nos relatórios.
O Dockerfile gera ambos os executáveis e assets em imagem sem root. Para Railway,
configurar `railway.web.toml` e `railway.engine.toml` em serviços separados.
Scripts de deploy e backup em `scripts/`; detalhes de recuperação e variáveis em
[docs/deploy-checklist.md](docs/deploy-checklist.md).

Histórico/dados externos, integrações autenticadas, deploy real e validação
quantitativa prolongada dependem da configuração do proprietário. O CI e os
fixtures comprovam comportamento técnico, não desempenho financeiro.
