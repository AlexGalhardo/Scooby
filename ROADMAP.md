# ROADMAP — Scooby

**Documento principal de onboarding e continuidade do projeto.**
Atualizado em 2026-11-03. Ler antes de planejar ou alterar o código.
Repositório: [AlexGalhardo/Scooby](https://github.com/AlexGalhardo/Scooby).

## 1. Ponto de retomada

O projeto está na **fase de especificação**, com starter funcional e documentação
versionada. A plataforma de trading ainda não foi implementada. A próxima ação
é continuar a **rodada 3 de decisões (Q7–Q9)** em [docs/grilling.md](docs/grilling.md),
registrar as respostas e resolver as dependências necessárias ao primeiro
incremento. Não interpretar uma recomendação como aceite do proprietário.

A sessão anterior entregou planejamento, pesquisa web, 6 SDDs e 10 ADRs.
Não recriar o projeto, não trocar a stack e não repetir perguntas já respondidas.
Nenhum deploy Railway, projeto OAuth ou integração autenticada com corretora
foi criado por essas entregas.

## 2. Onboarding para uma nova sessão ou agente

1. Executar `git status --short --branch` e `git log -5 --oneline`; preservar
   alterações preexistentes. Ler eventuais `AGENTS.md` aplicáveis no checkout.
2. Ler este arquivo para entender o estado e o próximo passo.
3. Ler [produto](docs/produto.md) e [grilling](docs/grilling.md), distinguindo
   requisitos confirmados, propostas e perguntas pendentes.
4. Ler os [ADRs](docs/adrs/README.md), a [arquitetura](docs/arquitetura.md) e o
   SDD da entrega em que vai trabalhar. Não tratar ADR proposto como aceito.
5. Consultar [pesquisa web](docs/web-research.nd) e [fontes](docs/fontes.md) antes
   de depender de uma capacidade, tarifa ou contrato externo. Revalidar o que
   pode ter mudado; o relatório é datado, não garantia permanente.
6. Conferir o código e as ferramentas realmente disponíveis. Processos, arquivos
   em `/tmp`, caches e credenciais de uma sessão anterior podem não existir.
7. Identificar a menor entrega autorizada com critérios de aceite claros. Fazer
   perguntas sobre decisões do usuário; pesquisar fatos técnicos por conta própria.
8. Ao concluir, atualizar o checklist, evidências e ponto de retomada neste arquivo,
   além dos SDDs/ADRs afetados. Deixar a próxima ação concreta e os bloqueios explícitos.

Instruções atuais do usuário prevalecem sobre recomendações anteriores. Se houver
contradição entre documentos, conferir decisão original/registro de grilling e
código; corrigir a divergência, sem inventar uma resposta.

## 3. Produto e decisões confirmadas

Scooby é um terminal **pessoal** para criar, testar, comparar e acompanhar
estratégias quantitativas, com execução simulada e controles de risco.

| Tema | Decisão confirmada |
| --- | --- |
| Público | Uso pessoal no MVP; sem SaaS ou cadastro público |
| Mercados | Brasil/B3 e ações americanas; ambos fazem parte do escopo |
| Execução | Backtesting e paper trading; capital real fora do MVP |
| EUA | Primeira integração com Alpaca Paper |
| B3 | Simulador interno inicialmente; broker externo depende de API comprovada |
| Identidade | Uma conta Google autorizada; email verificado no bootstrap e vínculo ao subject estável |
| Entrada | `/` é o dashboard principal autenticado |
| Visitante | Redirect automático para `/login`; não para o Google diretamente |
| Login | `/login` oferece somente o botão Google; sucesso retorna a `/` |
| Interface | Dashboard com sidebar usando shadcn/ui |
| Stack | Go + Vukem, Inertia.js + React + TypeScript, Tailwind CSS, PostgreSQL |
| Hospedagem | Railway |
| Continuidade | Engine deve funcionar sem depender de aba ou sessão do navegador |

Detalhes de topologia, outbox, cache e mensageria continuam **propostos**, conforme
os ADRs. TimescaleDB, Redis, NATS, Python de pesquisa e LLM não são dependências
obrigatórias já aprovadas. Mercado real não é ativado por simples troca de URL.

## 4. Estado real do repositório

| Componente | Estado observado |
| --- | --- |
| Starter oficial Vukem | Implementado, com home e CRUD público de posts |
| Go / Vukem | Go 1.27.3; Vukem 0.37.0; driver PostgreSQL 0.6.1 |
| Frontend | Inertia React 3.9.1; React 19.3.1; Tailwind 4.3.3; Vite/TypeScript |
| PostgreSQL | Driver registrado, migration de exemplo e Compose |
| CI | Go/race/vet, tipos/build frontend, vulnerabilidades, segredos e imagem |
| Documentação | Produto, arquitetura, dados, operação, validação, pesquisa, SDDs e ADRs |
| Google-only e conta autorizada | Especificados; não implementados |
| Dashboard financeiro / shadcn/ui | Especificados; não implementados |
| Feed EUA/B3 e datasets | Não integrados/adquiridos/validados |
| Estratégias / backtest / risco | Não implementados |
| Worker / adapters / reconciliação | Não implementados |
| Deploy Railway | Planejado; não realizado |

`routes/web.go` ainda publica `/` e `/posts`. A home atual não é o dashboard
protegido descrito no produto. Arquivo de design ou ADR aceito não é evidência
de feature pronta. Não preencher o dashboard com números inventados apresentados
como posições ou lucros reais.

### Mapa do código existente

| Caminho | Uso |
| --- | --- |
| `cmd/app/main.go` | Entrada do servidor/CLI da aplicação |
| `bootstrap/app.go` | Providers, banco, migrations, Inertia e composição |
| `routes/web.go` | Rotas atuais e middleware |
| `app/controllers/` | Controllers do starter |
| `database/migrations/` | Schema existente; atualmente exemplo de posts |
| `resources/js/pages/` | Páginas React/Inertia |
| `resources/js/layouts/` | Layout atual a adaptar |
| `resources/css/app.css` | Estilos Tailwind |
| `resources/views/app.html` | Documento HTML e bootstrap Inertia |
| `tests/` | Testes do starter e integração PostgreSQL |
| `Dockerfile`, `compose.yaml` | Build e ambiente de desenvolvimento |
| `.github/workflows/ci.yml` | Verificações automatizadas |

`cmd/engine` e `internal/trading` são diretórios **futuros propostos**, não
componentes executáveis existentes. O Dockerfile atual só produz `/app/app`.

## 5. Decisões abertas e bloqueios

| Prioridade | Pendência | Impacto |
| --- | --- | --- |
| Agora | Q7: ações à vista, long-only, sem alavancagem e watchlist proposta | Instrumentos, precisão e tipos de operação |
| Agora | Q8: IEX EUA e CSV histórico B3/replay, ou dados B3 em tempo real | Feed, licença, histórico e orçamento |
| Agora | Q9: candles fechados de 5 minutos e sessão regular | Calendário, warmup, execução e encerramento |
| Depois do universo | Regras exatas das três estratégias candidatas | Indicadores, sinais e versões |
| Depois das estratégias | Limites numéricos, stops, sizing, perda diária e emergência | Motor de risco e critérios de aceite |
| Antes da autenticação/deploy | Recuperação do proprietário e vínculo Google | Evitar perda de acesso e cadastro indevido |
| Antes das integrações | Contas, permissões e segredos por ambiente | OAuth, Alpaca Paper, Railway e dados B3 |
| Antes de operar continuamente | RPO/RTO, retenção, alertas, responsável e orçamento | Recuperação e monitoramento |
| Antes de implementar arquitetura proposta | Revisar/aceitar ADRs necessários à entrega | Worker, persistência, eventos e extensões |

PETR4, VALE3, ITUB4, AAPL, NVDA e TSLA, IEX/CSV e cinco minutos são **propostas
sem resposta**, não configuração final. Não repetir Q1–Q6: suas respostas já
estão registradas. Nunca pedir valores de segredos no chat ou escrevê-los no Git.

## 6. Roadmap com critérios de conclusão

Legenda: `[x]` artefato entregue; `[ ]` trabalho ainda pendente. Uma etapa só
muda para concluída com evidência de código/verificação, não porque foi descrita.

### P0 — Fundação e especificação — em andamento

- [x] Gerar starter oficial e configurar origin `AlexGalhardo/Scooby`.
- [x] Configurar PostgreSQL, Docker Compose, lockfiles e CI inicial.
- [x] Criar produto, brainstorming, arquitetura, modelo de dados e plano.
- [x] Criar 6 SDDs e 10 ADRs com estados explícitos.
- [x] Registrar respostas das rodadas 1 e 2.
- [x] Pesquisar fontes oficiais e documentar em `docs/web-research.nd`.
- [x] Criar este documento principal de onboarding e continuidade.
- [ ] Obter respostas Q7–Q9 e continuar somente pelas decisões desbloqueadas.
- [ ] Definir regras, limites e dependências externas necessárias ao primeiro incremento.
- [ ] Registrar revisão conjunta e aceite do desenho necessário à implementação.

**Saída:** escopo acionável, decisões necessárias resolvidas e critérios de aceite
revisáveis. Não esperar adquirir todos os serviços para fazer trabalho independente.

### P1 — Identidade e shell do dashboard — não iniciado

- [ ] Implementar Google OIDC e gate da conta única antes de criar sessão.
- [ ] Persistir vínculo por subject, sessão e proteção contra replay OAuth.
- [ ] Tornar `/` protegido; `/login` com único botão Google; logout com CSRF.
- [ ] Desativar rotas de senha/cadastro/reset e remover/proteger exemplos públicos.
- [ ] Instalar shadcn/ui, alias para `resources/js` e sidebar responsiva.
- [ ] Implementar estados vazios, carregamento, erro e ausência de dados.

**Aceite:** [SDD-01](docs/sdd/01-identidade-dashboard.md); visitante não recebe
dados privados, outro Google não entra, callback inválido falha, logout revoga
sessão e navegação/teclado/mobile funcionam. Não depende de inventar dados de mercado.

### P2 — Dados financeiros e ingestão — não iniciado

- [ ] Criar instrumentos, contas/modos, candles, procedência e versões do schema.
- [ ] Definir decimal, tick/lote, moedas, UTC e calendários por mercado.
- [ ] Implementar adapter de dados EUA com feed autorizado e histórico coerente.
- [ ] Implementar entrada B3 conforme Q8; validar granularidade/licença e manifesto.
- [ ] Detectar dados duplicados, gaps, revisões, outliers e atraso.

**Aceite:** [SDD-02](docs/sdd/02-dados-estrategias.md),
[SDD-06](docs/sdd/06-adapters-simulacao.md) e
[modelo de dados](docs/modelo-de-dados.md). Não fabricar candles intraday de OHLC
diário; dados sem fonte validada não demonstram robustez de estratégia.

### P3 — Estratégias e backtesting — não iniciado

- [ ] Especificar regras e parâmetros das estratégias aceitas pelo proprietário.
- [ ] Implementar funções determinísticas, warmup e configuração versionada.
- [ ] Criar replay e broker simulado com hipóteses explícitas de fill/custos.
- [ ] Versionar dataset, código, parâmetros e relatórios.
- [ ] Separar desenvolvimento, validação e teste temporal intocado.

**Aceite:** [SDD-04](docs/sdd/04-backtesting.md); resultado reproduzível,
sem look-ahead, custo/liquidez documentados e métricas líquidas com limitações.

### P4 — Risco, execução paper e recuperação — não iniciado

- [ ] Implementar política de risco e reservas transacionais por conta.
- [ ] Persistir sinais, intenções, outbox, ordens, fills e ledger idempotentes.
- [ ] Criar worker independente da web, liderança e fencing conforme desenho aceito.
- [ ] Implementar Alpaca Paper EUA e simulador B3 com modos claramente separados.
- [ ] Implementar reconciliação no boot/periódica e tratamento de `unknown`.
- [ ] Persistir bloqueio de emergência; distinguir bloquear, cancelar e zerar.

**Aceite:** [SDD-03](docs/sdd/03-risco-execucao.md), SDD-06 e matriz de falhas em
[validacao.md](docs/validacao.md). Testar timeout após aceite externo, fills
parciais/duplicados, crash, feed vencido, deploy sobreposto e posições divergentes.
Fechar navegador não interrompe a engine.

### P5 — Operação e publicação Railway — não iniciado

- [ ] Alimentar dashboard com snapshots autorizados, modo, moeda e idade dos dados.
- [ ] Implementar auditoria e atualização autenticada sem expor segredos.
- [ ] Preparar binários/imagens e configuração de serviços web/engine/PostgreSQL.
- [ ] Configurar migrations serializadas, readiness, heartbeat e alertas.
- [ ] Ensaiar restore, restart, SIGTERM e incidentes em ambiente de teste.
- [ ] Publicar na Railway quando essa etapa estiver autorizada, com dependências prontas.

**Aceite:** [SDD-05](docs/sdd/05-operacao-railway.md) e
[runbook Railway](docs/railway.md). Engine sem sleeping; healthcheck de deploy não
substitui monitoramento contínuo. Nenhum binário futuro é configurado antes de existir.

### P6 — Validação contínua — não iniciado

- [ ] Observar paper/backtest nos cenários e períodos acordados.
- [ ] Comparar desempenho e custos sem ajustar sobre o conjunto final.
- [ ] Revisar drawdown, divergências, alertas, dados e limites operacionais.
- [ ] Registrar evidências, falhas e decisão de manutenção/desativação de estratégia.

**Aceite:** critérios numéricos e amostrais definidos com o proprietário;
[plano de validação](docs/validacao.md). Paper não comprova lucro futuro.

**Fora do MVP:** P7/live, SaaS, sinais pagos, derivativos, ML e apuração fiscal.
Uma fase futura exige novo escopo e critérios próprios; este roadmap não autoriza
ordens com capital real.

## 7. Executar e verificar a base

Requisitos: Go 1.27.3+, Node.js 24+, npm e Docker Compose. Confirmar versões
fixadas em `go.mod`, `package.json` e lockfile antes de instalar. Não substituir
um `.env` existente nem regenerar `APP_KEY` de uma instalação em uso.

Em um clone novo, criar `.env` a partir de `.env.example` somente se ausente;
instalar a CLI Vukem 0.37.0 e gerar a chave inicial. Instruções completas no
[README](README.md). Depois:

```sh
npm ci
docker compose up -d postgres
vukem migrate
vukem serve
```

Sem CLI, `go run ./cmd/app migrate` e `go run ./cmd/app` executam os comandos
Go; `npm run dev` precisa de outro terminal para Vite. Porta padrão `3000` no
ambiente de execução não é um endereço público de deploy.

Verificações conforme a alteração:

```sh
npm run typecheck
npm run build
go test -race ./...
go vet ./...
go tool govulncheck ./...
git diff --check
```

Para integração PostgreSQL, usar banco dedicado e `TEST_DATABASE_URL`, conforme
README. A suíte isolada do starter usa SQLite em memória; isso não muda o banco
principal PostgreSQL. Para documentação apenas, checar links locais, consistência,
whitespace e segredos; não exigir todos os testes de runtime sem mudança de código.

Em ambiente gerenciado, ler as instruções/skill de cloud disponíveis antes de
rede ou Docker. Respeitar proxy/CA e permissões. Ferramentas instaladas em
`/workspace/.tools` são auxiliares desta sessão, não dependências versionadas.
Não usar o eventual `/usr/bin/go` sem conferir `go version`: nesta preparação
ele não era o compilador Go. Não despejar variáveis de ambiente ou arquivos secretos.

## 8. Invariantes que as implementações devem preservar

- Autenticação Google não dá permissão de operar uma corretora; autorização é local.
- Conta única não elimina guards de servidor em páginas, JSON, streams e comandos.
- Separar BRL/USD, papel/simulação/replay/backtest e suas fontes/contas/ledgers.
- Toda intenção passa pelo risco; reserva inclui ordens abertas, parciais e ambíguas.
- Timeout não prova rejeição; persistir ID, consultar e reconciliar antes de retry.
- Fill duplicado não pode alterar saldo duas vezes. Cancelamento pendente não é flat.
- Falha de banco/feed ou divergência bloqueia novas exposições conforme política.
- Lock/fencing não desfaz request externo já em voo; tratar recuperação explicitamente.
- Fechar sessão/browser não controla o ciclo de vida da engine.
- Logs, props Inertia e bundles nunca contêm credenciais de broker/OAuth.
- Revisão de código ou deploy não reverte operações; ledger preserva ajustes/auditoria.

## 9. Mapa documental e fonte de cada tipo de verdade

| Documento | Quando consultar/manter |
| --- | --- |
| **Este ROADMAP.md** | Onboarding, checklist, estado e próxima ação |
| [docs/README.md](docs/README.md) | Catálogo dos documentos e estados |
| [docs/produto.md](docs/produto.md) | Requisitos RF/RP e escopo |
| [docs/grilling.md](docs/grilling.md) | Respostas do proprietário e fronteira de perguntas |
| [docs/brainstorming.md](docs/brainstorming.md) | Alternativas e hipóteses, não aceites |
| [docs/arquitetura.md](docs/arquitetura.md) | Fluxo e responsabilidades propostos |
| [docs/modelo-de-dados.md](docs/modelo-de-dados.md) | Precisão, integridade e entidades |
| [docs/planejamento.md](docs/planejamento.md) | Dependências e rastreabilidade detalhadas |
| [docs/sdd/README.md](docs/sdd/README.md) | Contratos e critérios de aceite por módulo |
| [docs/adrs/README.md](docs/adrs/README.md) | Decisões aceitas/propostas e consequências |
| [docs/web-research.nd](docs/web-research.nd) | Pesquisa datada, fontes e limitações verificadas |
| [docs/railway.md](docs/railway.md) | Configurações e runbooks de operação |
| [docs/validacao.md](docs/validacao.md) | Testes técnicos, quantitativos e de falha |

## 10. Registro de continuidade e evidências

| Marco | Referência | Evidência e limite |
| --- | --- | --- |
| Fundação | `205f7a3`, `9e21d88` | Starter; testes/typecheck/build e integração PostgreSQL passaram na sessão de criação; imagem passou no Actions; não são evidência da futura engine |
| Especificações | `4e2adde` | 28 documentos, 6 SDDs, 10 ADRs; decisões Q1–Q6 registradas |
| Pesquisa web | `4664823` | 24 URLs, links internos e scan de segredos verificados; sem integração autenticada externa |
| Onboarding | Esta alteração | ROADMAP principal e links de entrada; links locais, whitespace e scan de segredos verificados |

Não afirmar que CI continua verde com base em execução antiga. Reexecutar checks
pertinentes ou consultar a execução correspondente ao commit em questão.
Não registrar o próprio hash antecipadamente; usar o histórico Git para localizá-lo.

### Atualização obrigatória ao encerrar trabalho

Registrar neste arquivo, de forma curta:

- entrega concluída e evidência (arquivo, teste, commit/PR ou resultado verificável);
- escolhas novas e quais SDDs/ADRs foram atualizados;
- limitações/bloqueios que continuam abertos;
- primeira próxima ação e seus pré-requisitos;
- se houve mudança de estado operacional ou dependência externa.

Não marcar `[x]` para trabalho parcial. Não apagar decisões antigas para esconder
mudança de escopo. Usar commits revisáveis, sem segredos nem caches/builds gerados.

**Próxima ação ao retomar:** continuar Q7–Q9 em `docs/grilling.md`. As respostas
não foram recebidas até esta atualização. Depois, atualizar requisitos e SDDs e
resolver a próxima fronteira; iniciar P1 quando o incremento estiver autorizado
com suas decisões necessárias resolvidas.
