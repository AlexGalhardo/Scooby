# ROADMAP — Scooby

**Documento principal para onboarding e continuidade entre sessões/agentes.**
Atualizado: 2026-11-03. Repositório: [AlexGalhardo/Scooby](https://github.com/AlexGalhardo/Scooby).

## 1. Ponto de retomada

O terminal pessoal **paper-only está implementado**, incluindo web, worker,
backtesting, risco e adapters. Projeto e serviços Railway foram provisionados;
web e engine saudáveis, com web em https://scooby-web-production.up.railway.app. O próximo passo é
substituir placeholders por credenciais Google/Telegram/Alpaca e validar as
integrações, conforme [docs/deploy-checklist.md](docs/deploy-checklist.md).
Não recriar o starter nem repetir a entrevista de decisões já encerradas.

Templates por serviço disponíveis em `.env.railway.web.example` e
`.env.railway.engine.example`; cópias `.local` são privadas e ignoradas pelo Git.
O usuário está criando Google OAuth e Telegram. Não substituir credenciais reais
por placeholders nem considerar autenticação/publicação Railway concluídas sem
confirmação da CLI/API. Ver [checklist](docs/deploy-checklist.md#templates-de-variáveis).

A última instrução do usuário autorizou implementar todo o roadmap sem interrupções
e pedir configurações de deploy ao final. Isso substitui o bloqueio anterior de
Q7–Q9: defaults foram adotados por delegação e documentados no ADR-0011, sem
inventar respostas individuais. Estratégias/contas começam desabilitadas/bloqueadas.

**Não concluído por falta de acesso externo:** OAuth Google real, credenciais e
permissões Alpaca Paper, dados históricos B3 licenciados,
alertas reais e acompanhamento quantitativo prolongado. Não confundir testes
locais com operação/lucratividade comprovada. Capital real permanece fora do MVP.

## 2. Onboarding de uma nova sessão

1. Ler `git status --short --branch`, `git log -5 --oneline` e eventuais `AGENTS.md`;
   preservar alterações preexistentes.
2. Ler este ROADMAP, o [README](README.md) e o [checklist de deploy](docs/deploy-checklist.md).
3. Conferir os [ADRs](docs/adrs/README.md), especialmente 0011–0014, e o SDD da tarefa.
4. Revisar o código: checklist/documento não substitui evidência de implementação.
5. Consultar [pesquisa web](docs/web-research.nd); revalidar políticas, versões e
   calendários antes de ampliar cobertura. Ela contém limitações de acesso.
6. Inspecionar ferramentas e bindings de segredos sem imprimir valores. Processos,
   caches, `/tmp` e `/workspace/.tools` não são dependências versionadas garantidas.
7. Continuar a menor entrega autorizada, executar verificações pertinentes e
   atualizar este arquivo com evidências, pendências e próxima ação concreta.

Instruções atuais do usuário prevalecem. A skill grilling foi usada na fase de
planejamento; a ordem posterior de execução autônoma permite escolhas técnicas
rotineiras sem reabrir aprovação. Novos mercados, SaaS ou capital real exigem novo
escopo. Não solicitar tokens no chat: usar variáveis seguras do ambiente/provedor.

## 3. Escopo e escolhas

| Tema | Estado/decisão |
| --- | --- |
| Público | Pessoal; única conta Google autorizada; sem cadastro público |
| Mercados | B3 + ações EUA; BRL/USD segregados |
| Execução | Alpaca Paper EUA; replay interno B3; nenhum adapter live |
| Login | `/` protegido → visitante vai a `/login`; único botão Google; sucesso em `/` |
| Interface | React/Inertia, Tailwind, shadcn/ui Radix, sidebar e gráficos |
| Backend | Go/Vukem, PostgreSQL; web e engine em processos distintos |
| Estratégias | SMA crossover, reversão Z-score, rompimento; versões imutáveis |
| Universo inicial | PETR4, VALE3, ITUB4, AAPL, NVDA, TSLA; long-only, sem leverage |
| Dados | Candles fechados de 5m; IEX EUA e CSV autorizado B3 |
| B3 | Calendário revisado 2026, feriados/abertura especial e grade por período |
| EUA | Calendário da API Alpaca + clock antes de operar |
| Risco inicial | 50 bps/operação; perda diária 100 bps; exposição 5000 bps; 3 posições |
| Simulação | Custos 5 bps/lado, slippage 5 bps; editáveis com conta bloqueada |
| Infra | Railway; web/engine/PostgreSQL; sem dependência inicial Redis/NATS/TimescaleDB |

Defaults de ativos/feed/timeframe/limites foram adotados na implementação delegada,
não tratados como recomendações financeiras. B3 começa com patrimônio fictício
identificado; EUA não mostra patrimônio conhecido antes de reconciliar a conta.

## 4. Estado do código e caminhos relevantes

| Caminho | Implementação |
| --- | --- |
| `cmd/app`, `bootstrap` | Servidor/CLI Vukem; migrations, providers e defaults |
| `cmd/engine` | Processo contínuo paper; não depende da sessão web |
| `routes/web.go`, `routes/social.go` | Rotas privadas, Google-only e APIs |
| `app/accounts` | Gate de email verificado, vínculo por subject e sessões |
| `app/controllers/terminal` | Snapshots, importação, estratégias, jobs e controles |
| `app/commands` | Worker/readiness e recuperação de proprietário por CLI |
| `internal/trading/domain.go` | Decimais, sizing, reservas, fills, ledger e políticas |
| `internal/trading/calendar.go` | Calendário B3 e limites de sessão |
| `internal/trading/data.go` | CSV estrito, validação, hashes e gaps |
| `internal/trading/backtest.go` | Replay determinístico, custos, curvas e métricas |
| `internal/trading/store.go` | Transações, audit, versões, jobs, artefatos e snapshots |
| `internal/adapters` | Alpaca paper/data/calendar e alertas Telegram opcionais |
| `internal/engine` | Liderança, fencing, replay, dispatch e reconciliação |
| `database/migrations` | Identidade, sessão/cache, contas, datasets, estratégias/jobs/audit |
| `resources/js/pages/Terminal.tsx` | Visão geral, mercado, estratégias, backtests, ordens, posições, risco, auditoria, settings |
| `resources/js/components/ui` | Componentes gerados shadcn/ui |
| `resources/js/pages/Auth/Login.tsx` | Login Google exclusivo |
| `tests`, `internal/*/*_test.go` | Auth, domínio, adapters, falhas, concorrência e failover |
| `cmd/e2e`, `e2e` | Fake OAuth apenas KindTest/build tag e teste browser |
| `Dockerfile`, `railway.*.toml`, `scripts` | Dois binários, deploy sequencial e backup |

Stack fixada: Go 1.27.3; Vukem 0.37.0; driver PostgreSQL 0.6.1; Inertia React
3.9.1; React 19.3.1; Tailwind 4.3.3; Vite/TypeScript e dependências no lockfile.
`/posts`, cadastro local, senha e reset de senha não estão publicados.

## 5. Checklist de entregas

`[x]` significa código/artefato entregue. Validação externa permanece separada.

### P0 — Fundação e especificação

- [x] Starter oficial, PostgreSQL, Compose, Git e CI.
- [x] Produto, arquitetura, modelo lógico, planejamento, pesquisa web e onboarding.
- [x] 6 SDDs, ADRs originais e ADRs 0011–0013 de implementação.
- [x] Q1–Q6 registradas; defaults Q7–Q9 adotados sob autorização posterior.
- [x] Regras/limites/configuração e procedimento de recuperação definidos.

### P1 — Identidade e dashboard

- [x] Login Google-only; conta única; subject persistente; state/nonce/PKCE nativos.
- [x] Sessões e cache compartilhados; rotação/revogação; CSRF e throttle OAuth.
- [x] `/` privado, `/login`, logout; sem superfícies locais de senha/cadastro.
- [x] Sidebar shadcn/ui, tema, mobile, gráficos e estados vazios/erro/loading.
- [x] Guards de servidor em páginas e APIs; nenhuma credencial nas props/bundle.
- [ ] Exercitar OAuth Google real com domínio/callback finais.

### P2 — Dados financeiros

- [x] Contas/modos/moedas segregados e instrumentos com tick/lote.
- [x] CSV atômico com procedência/licença/hash, OHLCV, offsets e grade de 5m.
- [x] Rejeição de entradas inválidas/duplicadas/fora de ordem; gaps identificados.
- [x] Histórico IEX via adapter e calendário/clock Alpaca.
- [x] B3 2026: feriados, Cinzas 13h, grade março/novembro; anos não revisados rejeitados.
- [ ] Receber e validar dataset B3 intraday efetivamente licenciado.
- [ ] Validar IEX com permissões e limites da conta escolhida.

### P3 — Estratégias e backtesting

- [x] Médias móveis, reversão à média e rompimento determinísticos, com warmup.
- [x] Configurações imutáveis; nova versão numerada; ativação separada da conta.
- [x] Jobs persistentes, cancelamento, recuperação e timeout de processamento.
- [x] Próxima barra elegível, custos, slippage, participação máxima por volume/lote.
- [x] Stop conservador e gap-through; relatório identifica hipóteses e limitações.
- [x] Dataset/config/código identificados; curva, benchmark bruto e artefato JSON.
- [x] Métricas de retorno, custos, drawdown/duração, turnover, trades e ganhos/perdas.
- [x] Partições temporais 70/30 e modo exploratório explicitamente identificado.
- [ ] Avaliar estratégias em amostras licenciadas/out-of-sample significativas.

### P4 — Risco, paper e recuperação

- [x] Decimal, sizing, limites de caixa/exposição/perda diária e posições.
- [x] Reserva + intenção/outbox no mesmo snapshot/transação e lock por conta.
- [x] Fills deduplicados, parciais, custos, ledger e P&L de posições.
- [x] Worker independente; advisory lock dedicado, standby e leader_epoch.
- [x] Reconciliação antes de novas entradas; timeout sem retry cego; estado unknown.
- [x] Bloquear, cancelar e solicitar zeragem são ações distintas e auditadas.
- [x] Proteção ao desabilitar estratégia com risco pendente; B3 usa um replay ativo por conta.
- [x] Testes de reservas concorrentes, rollback, timeout/partial/dedupe e failover.
- [ ] Exercitar contratos/cancelamentos/fills na conta Alpaca Paper dedicada real.

### P5 — Railway e operação

- [x] Dashboard conectado aos snapshots PostgreSQL com polling autenticado.
- [x] Auditoria de ações/políticas/rejeições, heartbeat e readiness da engine.
- [x] Dois executáveis e imagem sem root; configuração API web/engine e Compose.
- [x] Deploy sequencial, healthcheck, migrations e script de backup.
- [x] Recuperação administrativa de proprietário auditada, com revogação de sessões.
- [x] Backup/restore ensaiado em PostgreSQL de teste isolado.
- [x] CLI Railway autenticada por OAuth; projeto, serviços, variáveis e domínio criados.
- [x] Web publicada; build, migrations, login HTTP e acesso ao banco confirmados.
- [x] Configuração pela API atual e Dockerfile portável (ADR-0014), sem TOML legado.
- [x] Web, engine e PostgreSQL publicados e saudáveis no domínio Railway.
- [x] Backup diário PostgreSQL agendado; retenção do provedor confirmada em seis dias.
- [ ] Confirmar orçamento/adequação regional e configurar monitoramento externo.
- [ ] Validar primeira execução do backup e complementar retenção se exigir sete dias.
- [ ] Testar Google/broker/Telegram e restore no provedor escolhido.

### P6 — Validação contínua

- [x] Instrumentação e relatórios para acompanhamento e revisão de drawdown/custos.
- [ ] Observar paper ao longo dos períodos/cenários acordados; registrar divergências.
- [ ] Comparar resultados fora da amostra sem otimizar no teste final.
- [ ] Medir latência, capacidade, RPO/RTO e revisar políticas com evidências externas.
- [ ] Registrar decisão de manutenção/desativação por estratégia.

P6 depende de passagem do tempo, dados e contas configuradas; não marcar como
concluído por fixtures sintéticos. P7/live, SaaS, derivados, ML e apuração fiscal
não foram adicionados ao escopo.

## 6. Contratos e limitações importantes

- Persistência financeira usa JSON decimal versionado em conta PostgreSQL, com
  lock/CAS e audit no mesmo commit; não alegar tabelas NUMERIC normalizadas onde
  não existem. ADR-0012 explica tradeoffs e migração futura para volume alto.
- Outbox é a lista durável de ordens aprovadas no snapshot. Não há exactly-once
  externo; request já em voo exige reconciliação mesmo com geração/fencing.
- Polling web 10s; loop de execução 15s; candles 5m. Não é HFT ou stop nativo
  garantido na corretora. Alertas opcionais precisam de token/chat configurados.
- Calendário B3 suporta 2026. Ampliar anos somente com calendário revisado e testes.
  IEX não equivale ao mercado consolidado; CSV público diário não fabrica intraday.
- Snapshot financeiro cresce com o histórico. Não truncar ledger para economizar
  espaço; normalizar/arquivar com trilha antes de volume elevado.
- Somente uma configuração por ativo EUA e um replay B3 por conta ficam ativos.
  Desligar proteção com posição/ordem pendente é rejeitado; primeiro reduzir risco.
- B3 é replay histórico interno, não corretora B3 nem feed ao vivo. Os stops/alvos
  são executados pelo simulador/engine, com limites explícitos de liquidez/latência.
- A web pode usar chaves paper para importar IEX; todos os valores ficam no servidor.
  Não colocar credenciais em variáveis `VITE_*`, props, logs ou arquivos Git.

## 7. Executar e verificar

Instruções completas no README; não sobrescrever `.env` nem regenerar chave em uso.

```sh
npm ci
docker compose up -d postgres
vukem migrate
vukem serve
# Outro terminal:
ENGINE_HEALTH_PORT=3001 go run ./cmd/engine
```

```sh
npm run typecheck
npm run build
go test -race ./...
go vet ./...
go tool govulncheck ./...
npm run test:e2e
git diff --check
```

PostgreSQL integration requer `TEST_DATABASE_URL` para banco dedicado `*_test`.
E2E requer Chromium/Playwright e Go no PATH; usa fake OAuth apenas em KindTest,
build tag `e2e` e endereço loopback. Em Codex, `CHROMIUM_PATH=/usr/bin/chromium`.
Não compilar Go simultaneamente com Vite apagando/recriando assets embutidos.

O ambiente inicial tem Go válido em `/workspace/.tools/go/bin`; verificar antes
de reutilizar. Seguir instruções de cloud/proxy/CA disponíveis. O disco do executor
é limitado: build Docker local pode esgotar espaço; usar CI para validar a imagem,
sem alterar storage driver/proxy nem desabilitar TLS.

## 8. Evidências e continuidade

Verificações realizadas durante implementação:

- Auth/gate/replay de callback/logout, rotas privadas e validações de API.
- Domínio: decimal, sizing, perda diária, duplicatas, fills parciais, CSV e calendários.
- Adapter/engine: timeout após envio sem reenvio, partial/dedup e reconciliação.
- PostgreSQL real: concorrência de 16 reservas, rollback e troca de liderança.
- Browser: Google fake → dashboard → CSV → estratégia → backtest/relatório →
  navegação de todas as áreas → mobile → logout.
- Dump/restore em base isolada; readiness worker; compilação Go e frontend.
- Railway real: web `266b5d5c-d668-4176-9bdc-7748ce9fee58` e engine
  `80a0f3e2-9559-4269-a83b-facc46502e5c` com status SUCCESS no commit `4337761`.
  `/` 302, `/login` 200, `/ready` 200 e `/api/snapshot` sem sessão 401.
- CI do ajuste Docker/API aprovado:
  https://github.com/AlexGalhardo/Scooby/actions/runs/41655709136.
- CI e scan de segredos devem ser conferidos no commit final; não inferir estado
  atual apenas de execução antiga. Login/broker reais não foram exercitados.

Histórico: `205f7a3`/`9e21d88` fundação; `4e2adde` especificações;
`4664823` pesquisa; `daeed47` onboarding inicial. O commit de implementação será
localizado por `git log`; não registrar um hash futuro antes do commit existir.

Ao concluir cada sessão, atualizar checklist e registrar: mudança, teste/evidência,
ADRs afetados, bloqueios externos e primeira próxima ação. Preservar o histórico
de respostas; nunca marcar feature pronta por haver apenas documentação.

**Próxima ação:** no serviço web, substituir Google OAuth e email autorizado;
validar callback e login real. No serviço engine, configurar Telegram opcional e
Alpaca Paper; repetir chaves paper na web se usar importação IEX. Railway já está
autenticado/provisionado. Contas permanecem bloqueadas até revisão e retomada
explícita. Conferir [destinos](docs/deploy-checklist.md#destino-provisionado).
