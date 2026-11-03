# Publicação Railway — configurações e handoff

Estado: código implementado; publicação depende de credenciais do proprietário.
Não colar tokens ou client secrets na conversa ou em arquivos versionados.

## Informações não secretas para a publicação

Projeto/workspace Railway e ambiente desejado; nomes/IDs dos serviços; região e
plano/orçamento aceito; domínio próprio ou domínio Railway; email Google do
proprietário. Se não existir projeto, criar um projeto Scooby e emitir um token
restrito ao ambiente, em vez de fornecer acesso global à conta.

Serviços: `scooby-web`, `scooby-engine`, PostgreSQL. Config File Path de cada
aplicação: `railway.web.toml` e `railway.engine.toml`. Uma réplica para a engine,
Serverless/sleeping desabilitado. O worker continua processando sem o navegador.
Usar URL privada do PostgreSQL entre serviços. Apenas a web precisa de domínio
público; `/ready` da engine funciona para healthcheck na porta configurada.

## Segredos necessários, por destino

| Variável | Destino | Como obter |
| --- | --- | --- |
| `RAILWAY_TOKEN` | Ambiente seguro do Codex/CLI, não app | Token de projeto/ambiente Railway |
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

1. Provisionar PostgreSQL e backup. Configurar variáveis e caminho do TOML.
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
aguarda status de sucesso por deployment ID. Não cria projetos sem informação
sobre o destino. Não foi executado sem token.

## Backup e recuperação

`scripts/backup.sh` produz dump custom, com permissões restritas. Definir
`BACKUP_FILE` fora do repositório e usar armazenamento privado cifrado. Agendar
backup diário e retenção inicial de sete dias no provedor, confirmando o plano.
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
