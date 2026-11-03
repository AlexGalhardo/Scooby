# Operação e deploy na Railway

> Atualização da publicação: projeto provisionado, serviços criados e web online.
> Destinos, domínio e configuração efetiva estão no
> [checklist](deploy-checklist.md#destino-provisionado) e
> [ADR-0014](adrs/0014-railway-api-e-docker-portavel.md). O texto histórico abaixo
> não descreve o estado atual e não deve orientar configurações de variáveis.

> Atualização de implementação: consulte o [ROADMAP](../ROADMAP.md),
> [ADR-0011](adrs/0011-implementacao-autonoma.md),
> [ADR-0012](adrs/0012-snapshots-transacionais.md) e
> [ADR-0013](adrs/0013-workers-polling-e-deploy.md) para o estado efetivo.
> O texto abaixo preserva requisitos/desenhos originais, não uma declaração de
> que todos os detalhes de infraestrutura ou validação externa já ocorreram.

Estado: destino confirmado; topologia e configurações abaixo propostas. Nenhum
serviço Railway foi criado nesta etapa.

## Topologia inicial

| Serviço | Processo | Exposição |
| --- | --- | --- |
| web | Binário Vukem atual, `/app/app` | HTTPS público; páginas e APIs autenticadas |
| engine | Futuro binário `/app/engine` | Sem domínio público; conexão ao broker/feed |
| PostgreSQL | Serviço persistente | Rede privada entre serviços |
| backtest | Job/processo separado se necessário | Sem disputa de recursos com execução |

Manter engine continuamente ativa e desabilitar hibernação/serverless nela.
Inicialmente uma réplica; locks e fencing continuam necessários durante rolling
deploy. Não escalar executor por simples alteração de réplicas. Cache/broker
adicionais entram só por ADR. Confirmar região, plano, CPU/RAM, orçamento e
backup antes do deploy, sem assumir TimescaleDB disponível.

## Build e comandos

O Dockerfile atual só compila `cmd/app`; criar `cmd/engine` exige alterar build e
imagem. Não configurar um start command inexistente. A web escuta em
`0.0.0.0:$PORT`; assets Vite são embutidos. Manter usuário sem privilégios.

Proposta de pre-deploy: `/app/app migrate`. Migrations devem ser compatíveis com
a versão antiga durante sobreposição e usar serialização/lock. Expandir schema,
publicar consumidores e só depois remover campos. Falha de migration impede
promoção. Não rodar migrations destrutivas no boot de cada worker.

Na Railway, start command de Docker sobrescreve ENTRYPOINT; informar o executável
completo. Validar configuração em staging, sem presumir expansão de shell. Registrar uma configuração por serviço quando implementado.

## Variáveis propostas

| Serviço | Variáveis | Observação |
| --- | --- | --- |
| web | APP_ENV, APP_DEBUG, APP_URL, HOST, PORT | production, debug falso, URL HTTPS canônica |
| web | APP_KEY, DATABASE_URL | Segredos Railway; chave persistente entre deploys |
| web | SOCIAL_GOOGLE_CLIENT_ID, SOCIAL_GOOGLE_CLIENT_SECRET | Nomes existentes do scaffold Vukem |
| web | CACHE_STORE, SESSION_DRIVER | Definir armazenamento compartilhado e migrations |
| engine | DATABASE_URL, TRADING_MODE | `TRADING_MODE` é contrato novo proposto |
| engine | BROKER_API_KEY, BROKER_API_SECRET | Nomes genéricos propostos; só paper no escopo recomendado |
| engine | BROKER_BASE_URL, MARKET_DATA_FEED | Validados contra allowlist do adapter/mode |
| engine | ENGINE_ACCOUNT_ID | Nova configuração proposta, escopo da liderança |
| alertas | TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID | Opcionais, após escolha do canal |

IDs de usuário autorizado e limites serão definidos após Q1. Não enviar segredos
a props Inertia, bundle Vite, logs ou documentos. Google callback deve coincidir
com URI registrada, incluindo protocolo, domínio e path. Login não requer tokens
de acesso Google permanentes se nenhuma API Google será usada.

## Saúde, telemetria e recuperação

`/up` existe no starter como liveness; não expor saldos ou configurações ali.
Propor `/ready` para web/dependências. A Railway usa healthcheck para deploy;
monitoramento contínuo precisa de ferramenta/alertas próprios. Worker reporta
heartbeat no banco e readiness interna; não depende de acesso público ao endpoint.

Métricas: idade do último candle, atraso do feed, processamento, latência do
broker, intenções unknown, divergências, reservas de risco, heartbeat, fila/outbox,
falhas de login e conexões. Logs estruturados com request/event/order IDs;
redigir tokens, cabeçalhos e perfis pessoais. Retenção e metas ainda pendentes.

Runbook de incidente: persistir bloqueio da conta → confirmar novas entradas
bloqueadas → consultar estado broker → cancelar conforme política → reconciliar →
verificar posições e exposição → registrar causa → retomar explicitamente.
Zerar posições é operação distinta, sujeita a mercado/limites; falha ou falta de
conectividade deve aparecer como pendência, nunca como confirmação de flat.

Runbook de restart: adquirir liderança → carregar bloqueios/versões → reconciliar
ordens e fills → reconstruir exposição → validar feed/calendário → habilitar só se
a política permitir. Restore é feito em ambiente isolado sem execução habilitada;
reconciliar contra broker antes de qualquer retomada. Definir RPO/RTO e ensaiar
restore; volume persistente sozinho não é estratégia de backup.

Rollback de código não reverte fills ou cancela ordens. Evitar downgrade de schema
financeiro. Worker deve reagir a SIGTERM parando novas entradas, preservando
intenções e finalizando checkpoint; o broker continua fonte de confirmação.

O pre-deploy roda em container separado, com variáveis e rede privada, mas sem
volumes montados. Não depender de escrita em disco nessa etapa. Ausência de sticky
sessions na distribuição de réplicas reforça sessions e states compartilhados.

## Ajuste após rodada 2

EUA: segredos exclusivos de Alpaca Paper e allowlist de host paper. B3: adapter
interno sem credenciais de ordens externas. Web: `OWNER_GOOGLE_EMAIL` é proposta
de configuração para o gate inicial da única conta autorizada. Uma vez vinculada,
a identidade Google persistida prevalece; recuperação deve ser auditada.
