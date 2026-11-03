# SDD-01 — identidade Google e dashboard

Estado: RF-01–RF-04 e uso pessoal confirmados. Admissão restrita a uma conta confirmada em Q6; recuperação ainda pendente.

## Contrato de navegação

| Rota | Visitante | Autenticado |
| --- | --- | --- |
| GET `/` | 302 para `/login`; Inertia usa redirect de navegação compatível | Renderiza `Dashboard` |
| GET `/login` | Renderiza `Auth/Login` com um botão Google | Redireciona para `/` |
| GET `/auth/google/redirect` | Inicia OAuth/OIDC com state, nonce e PKCE | Pode voltar a `/` |
| GET `/auth/google/callback` | Valida fluxo e autorização; cria sessão e vai a `/` | Tratar fluxo inválido sem trocar identidade silenciosamente |
| POST `/logout` | Não cria sessão | CSRF, invalida sessão e redireciona para `/login` |
| GET `/dashboard` | Mesmo guard | Alias opcional que redireciona a `/` |
| GET `/up` | Liveness mínimo | Mesmo comportamento |

`/posts` e formulários demo serão removidos ou protegidos durante substituição.
Rotas locais de senha, cadastro, reset e verificação por email não devem ficar
acessíveis só porque o gerador `install:auth --social google` as escreveu.
Registre apenas Google, sem provedor arbitrário escolhido pelo cliente.

## Identidade e sessão

Usar suporte `social.Google` do Vukem; não implementar OAuth manual sem motivo.
Validar assinatura e claims OIDC, issuer, audience, expiração, nonce, state e
uso único. Identidade estável é provider + subject, não email editável. Verificar
email quando usado na política. Não vincular contas existentes por email sem
política explícita; não aceitar um endereço autodeclarado como autorização.

Rotacionar sessão no login, cookies Secure/HttpOnly/SameSite apropriados, TTL e
revogação. Estado OAuth de uso único deve ser compartilhado entre instâncias.
Usar cache/sessão persistentes compatíveis com Vukem e migrations. Sessão vencida
em JSON/stream informa 401; navegação HTML leva a `/login`. CSRF em mutações.
Limitar tentativas e retornar mensagens sem código/token ou detalhes internos.

Como Q1 confirmou uso pessoal, cadastro implícito do scaffold precisa de gate: só identidade
preautorizada pode entrar; bootstrap controlado do proprietário e recuperação
seguem Q6: email verificado configurado no servidor, com vinculação ao subject
estável; a recuperação ainda será detalhada. SaaS está fora do MVP. Google válido não implica acesso permitido.

## Interface

Usar SidebarProvider/Sidebar/SidebarInset e componentes shadcn/ui para cards,
tabelas, dialogs, formulários e feedback. Navegação proposta: Visão geral,
Mercado, Estratégias, Backtests, Ordens, Posições, Risco, Auditoria e Configurações.
Mostrar apenas módulos entregues ou estado indisponível claro.

Cabeçalho: conta, mercado, PAPER/REAL, conectividade e atraso dos dados. Cards:
equity, P&L realizado/não realizado, exposição e drawdown. Conteúdo: gráfico,
ordens recentes e saúde da engine. Moeda, período, fonte e horário de atualização
são explícitos. P&L ausente não vira zero; simulado não aparece como real.

Somente props mínimas: usuário público, permissões e snapshots autorizados. Nunca
credenciais, tokens Google ou segredos do broker. Cálculos oficiais no backend.
Tema, idioma, período padrão e posições dos módulos aguardam rodada de UX.

## Aceite

Atender testes de auth/UI em `../validacao.md`. Entrar pelo Google retorna a `/`;
visitante não vê dashboard antes do redirect; logout revoga sessão; botão único
não introduz login por senha. SSR não é requisito confirmado. Sidebar responsiva
e acessível; ações do risco são autorizadas no servidor e auditadas.

## Notas verificadas para implementação

No frontend existente, alias `@` deve apontar para `resources/js`, e o CSS fica em
`resources/css/app.css`; não recriar um projeto `src/` paralelo. A documentação
shadcn/ui atual tem variantes de primitivas: escolher uma (Base UI, Radix ou outra
suportada) e manter consistência antes de copiar componentes. Essa escolha ainda
é detalhe técnico proposto. No Google, limitar escopos a `openid email profile`.

## Bootstrap da conta única — confirmado em Q6

Configuração inicial `OWNER_GOOGLE_EMAIL` (nome proposto) contém email autorizado,
normalizado segundo contrato restrito, sem heurísticas de pontos/aliases Gmail.
Callback exige email verificado e comparação com configuração antes da criação
da sessão. Primeira vinculação grava subject com transação e unicidade, impedindo
dois callbacks concorrentes de registrar proprietários distintos. Após vincular,
subject é a chave de acesso; mudança de email não cadastra um novo proprietário.
A política de atualização/recuperação do vínculo será escolhida antes do deploy.
Nenhuma tela pública solicita a configuração de proprietário. Segredos e email
real do proprietário serão configurados em ambiente, não neste documento.
