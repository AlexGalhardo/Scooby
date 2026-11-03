# Publicação Railway — configurações e handoff

Estado: projeto provisionado; web, engine e PostgreSQL saudáveis em 2026-11-03; integrações externas
aguardam credenciais reais. Ver destinos abaixo e confirmar saúde na CLI.
Não colar tokens ou client secrets na conversa ou em arquivos versionados.

## Destino provisionado

| Recurso | ID/endereço |
| --- | --- |
| Projeto Scooby | `706b6cfc-45e2-4bfc-af6b-45a7c401ed37` |
| Ambiente production | `a0036a02-8b91-4463-96b5-622ed1fe3805` |
| scooby-web | `4611a2b2-b810-4941-9eb9-279a13bc7982` |
| scooby-engine | `aa2f67bb-69b5-470f-a676-abbfc9d7a806` |
| Postgres | `17f09ea5-508e-4e1e-9c3c-40eeed4f3600` |
| Web | https://scooby-web-production.up.railway.app |
| Google callback | https://scooby-web-production.up.railway.app/auth/google/callback |

Projeto: https://railway.com/project/706b6cfc-45e2-4bfc-af6b-45a7c401ed37

Web e engine conectadas ao GitHub `AlexGalhardo/Scooby`, branch `main`.
Banco com referência privada; chave de aplicação gerada e compartilhada; valores
conferidos sem imprimi-los. Email do proprietário e OAuth permanecem placeholders;
Alpaca e Telegram estão vazios/desativados. O usuário está criando Google/Telegram.
Região inicial herdada do provedor: `us-east4-eqdc4a`, sem alteração de plano.
Deploys iniciais confirmados SUCCESS no commit `4337761`: web
`266b5d5c-d668-4176-9bdc-7748ce9fee58`; engine
`80a0f3e2-9559-4269-a83b-facc46502e5c`. Smoke HTTPS: `/` 302; `/login` 200;
`/ready` 200; `/api/snapshot` sem sessão 401. Engine healthcheck aprovado pelo
provedor; o contrato de readiness verifica heartbeat das duas contas no banco.

## Templates de variáveis

- `.env.example`: desenvolvimento; credenciais opcionais vazias com placeholders
  identificados em comentários, evitando chamadas externas com chaves fictícias.
- `.env.railway.web.example` e `.env.railway.engine.example`: templates por serviço.
- `.env.railway.web.local` e `.env.railway.engine.local`: cópias privadas ignoradas
  pelo Git, preparadas no workspace. A chave da aplicação é gerada uma vez e
  compartilhada entre as duas; não imprimir esses arquivos.

Os arquivos `.example` não são carregados automaticamente pela Railway. Aplicar
as variáveis no editor de cada serviço ou pela CLI autenticada. Não executar
`source` nos templates Railway: `${{Postgres.DATABASE_URL}}` é uma referência do
provedor, não uma expressão de shell. Ajustar `Postgres` ao nome real do serviço.
O domínio, email e OAuth têm marcadores explícitos que devem ser substituídos;
Alpaca e Telegram permanecem vazios até obter credenciais reais. O callback
Google fica no domínio da web, em `/auth/google/callback`.

Preparar arquivos locais não comprova que as variáveis foram aplicadas no provedor;
registrar projeto, serviços e domínio somente após confirmação da API Railway.

## Informações não secretas para a publicação

Projeto/workspace Railway e ambiente desejado; nomes/IDs dos serviços; região e
plano/orçamento aceito; domínio próprio ou domínio Railway; email Google do
proprietário. Se não existir projeto, criar um projeto Scooby e emitir um token
restrito ao ambiente, em vez de fornecer acesso global à conta.

Serviços: `scooby-web`, `scooby-engine`, PostgreSQL. Configuração via
`scripts/railway-configure.sh` e API atual; os TOMLs são referências históricas,
pois a API rejeitou Config as Code legado (ADR-0014). Uma réplica para a engine,
Serverless/sleeping desabilitado. O worker continua processando sem o navegador.
Usar URL privada do PostgreSQL entre serviços. Apenas a web precisa de domínio
público; `/ready` da engine funciona para healthcheck na porta configurada.

## Segredos necessários, por destino

| Variável | Destino | Como obter |
| --- | --- | --- |
| `RAILWAY_TOKEN` | Ambiente seguro do Codex/CLI, não app | Token de projeto/ambiente; sessão OAuth CLI também é aceita |
| `APP_KEY` | web e engine | Gerar uma vez com `vukem key:generate`; compartilhar a mesma chave e mantê-la |
| `DATABASE_URL` | web e engine | Referência ao PostgreSQL do projeto, preferencialmente rede privada |
| `SOCIAL_GOOGLE_CLIENT_SECRET` | somente web | Credencial OAuth tipo Web no Google Cloud |
| `ALPACA_API_KEY`, `ALPACA_API_SECRET` | somente engine | Chaves da conta Alpaca **Paper** |
| `TELEGRAM_BOT_TOKEN` | engine, opcional | Bot próprio para alertas operacionais |

A web atualmente também usa as chaves Alpaca para importar histórico a pedido do
proprietário. Para disponibilizar essa função, configurar as chaves paper na web
ou separar futuramente a ingestão em job. Nunca expor essas chaves em `VITE_*`.

Configurações adicionais:

- `APP_ENV=production`, `APP_DEBUG=false`, `APP_URL=https://DOMINIO-WEB`.
- `HOST=0.0.0.0`, `PORT` fornecido pela Railway; `ENGINE_HEALTH_PORT` vazio na
  Railway para usar `PORT` (localmente pode ser `3001`).
- `OWNER_GOOGLE_EMAIL`: email verificado admitido no bootstrap.
- `SOCIAL_GOOGLE_CLIENT_ID`: ID do cliente OAuth Web.
- Callback Google exato: `https://DOMINIO-WEB/auth/google/callback`.
- `SESSION_DRIVER=database`, `CACHE_STORE=database`, `MAIL_MAILER=log`.
- `TELEGRAM_CHAT_ID`: chat do proprietário, se optar por alertas.
- CLI: `RAILWAY_PROJECT_ID`, `RAILWAY_ENVIRONMENT`, `RAILWAY_WEB_SERVICE`,
  `RAILWAY_ENGINE_SERVICE` identificam o destino sem ambiguidade.

Configurar consentimento OAuth e usuário de teste se o app Google estiver em
modo Testing. Uma conta Google válida diferente da autorizada será rejeitada.
O email da configuração não substitui um subject já vinculado.

## Aplicar variáveis com segurança

Preferir o editor de variáveis Railway ou bindings secretos do ambiente Codex.
A CLI 5.71.0 foi instalada neste workspace e verificada pelo SHA256 publicado.
Para automação, `railway variable set NOME --stdin --skip-deploys` lê um único
valor da entrada; usar seletores explícitos de projeto/ambiente/serviço e não
imprimir o retorno ou arquivos contendo segredos. Evitar valores secretos em
argumentos de shell, screenshots e logs. Não iniciar login interativo por
supor que um token precise ser exibido; usar a autorização configurada.

## Ordem da publicação

1. Provisionar PostgreSQL e backup. Configurar variáveis e executar
   `scripts/railway-configure.sh` com IDs explícitos; não definir caminho de TOML.
2. Publicar web. Pre-deploy executa `/app/app migrate`; Vukem serializa migrations
   PostgreSQL com advisory lock. Configurar timeout do pre-deploy no serviço.
3. Esperar deploy bem-sucedido e `/ready`. `--ci` sozinho só acompanha build.
4. Publicar engine, esperar `/ready` e verificar heartbeat do dashboard.
5. Conferir Google real/callback/logout e admissão restrita.
6. Importar histórico B3 autorizado e/ou buscar IEX. Começar com backtest.
7. Revisar políticas e habilitar estratégia; contas começam **bloqueadas**.
   Retomar PAPER é ação explícita no painel. Não existe configuração live.
8. Exercitar cancelamento e reconciliação em conta paper dedicada antes de usar
   a conta paper pessoal. Validar alertas se contratados/configurados.

`scripts/railway-deploy.sh` publica sequencialmente serviços já configurados e
aguarda status de sucesso por deployment ID. Aceita token seguro ou sessão OAuth
CLI. Não cria projetos sem informação sobre o destino. A publicação inicial usou
GitHub source connect e a configuração API equivalente, sem token no chat.

## Backup e recuperação

Backup diário nativo habilitado e relido pela API para a instância de volume
`18f653a2-4123-4230-a47a-3d3a57d32ab6`. Agenda retornada: `56 19 * * *`, retenção
`518400` segundos (seis dias). A API não oferece retenção custom nesse contrato;
a meta inicial de sete dias exige complemento externo, se mantida. Agendamento
confirmado não significa que o primeiro backup ou restore já foi validado.

`scripts/backup.sh` produz dump custom, com permissões restritas. Definir
`BACKUP_FILE` fora do repositório e usar armazenamento privado cifrado. Agendar
backup externo quando necessário para complementar a retenção do provedor.
Meta operacional inicial: RPO 24h, RTO 2h; metas não foram medidas em produção.

Restore em banco isolado: `pg_restore --no-owner --no-acl` para uma base de teste;
nunca apontar engine/broker para o restore sem bloqueio e reconciliação. Foi
exercitado dump/restore de base PostgreSQL de teste, não recuperação de produção.

Recuperação de Google: operador autenticado no ambiente administrativo pode
executar `/app/app owner:reset --confirm RESET-OWNER --reason "motivo detalhado"`.
Isso revoga sessões (session_version), registra auditoria e reabre bootstrap com
`OWNER_GOOGLE_EMAIL`. Não exclui histórico financeiro e não é rota HTTP.
Manter o proprietário anterior até confirmar posse da nova conta. Nenhum bypass
de autenticação está incluído no binário de produção; `cmd/e2e` exige build tag.

## O que permanece dependente de evidência externa

Login Google real, permissões/dados/fills da conta Alpaca, disponibilidade
Railway, alertas reais, restauração no provedor, histórico B3 licenciado e
observação de estratégias por período significativo. Testes locais não comprovam
esses resultados nem desempenho financeiro. P6 permanece acompanhamento, não
uma caixa marcada por dados sintéticos de teste.

O ID externo da conta Alpaca fica vinculado após a primeira reconciliação. Rotação
de chaves deve permanecer na mesma conta paper; apontar para outra conta bloqueia
o executor, sem mesclar saldos, ordens e histórico. Migração de conta exige
procedimento administrativo próprio.
