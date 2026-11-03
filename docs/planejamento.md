# Plano de desenvolvimento

> Atualização de implementação: consulte o [ROADMAP](../ROADMAP.md),
> [ADR-0011](adrs/0011-implementacao-autonoma.md),
> [ADR-0012](adrs/0012-snapshots-transacionais.md) e
> [ADR-0013](adrs/0013-workers-polling-e-deploy.md) para o estado efetivo.
> O texto abaixo preserva requisitos/desenhos originais, não uma declaração de
> que todos os detalhes de infraestrutura ou validação externa já ocorreram.

Estado: rascunho. Não é ordem para implementar escolhas pendentes.

## Etapas e dependências

| Fase | Entrega verificável | Depende de | Gate de conclusão |
| --- | --- | --- | --- |
| P0 | Requisitos, SDDs, ADRs e grilling | Conversa | Q1–Q3 respondidas; completar rodadas dependentes e revisão conjunta |
| P1 | Google-only, `/`, sidebar shadcn/ui | P0 para autorização de usuários | Testes de sessão, callback e navegação; starter público removido |
| P2 | Schema financeiro e feed | Mercado, corretora e plano de dados | Precisão, calendário, gaps e revisão de candles testados |
| P3 | Estratégias e backtest | Dataset + regras aprovadas | Replay determinístico, custos e separação temporal |
| P4 | Worker, risco e paper | Adapter + política de risco | Testes de falha, duplicação, execução parcial e restart |
| P5 | Dashboard operacional e Railway | P1–P4 | Reconciliação, alertas, restore e runbooks exercitados |
| P6 | Validação contínua | P5 | Evidências fora da amostra e operação estável; critérios numéricos acordados |
| P7 | Possível live | Decisão explícita futura | Revisão própria, limites, monitoramento e rollback operacional |

Estimativas de semanas na conversa são ilustrativas. Não adotar cronograma antes
de definir escopo, acesso ao histórico, orçamento e critérios de aceite.

## Incrementos de implementação

1. Criar migrations de identidade e sessão; integrar Google pelo suporte nativo
   Vukem; restringir autorização conforme público escolhido.
2. Expor apenas `/login`, fluxo Google, assets e health públicos. Proteger páginas,
   JSON e streams. Home vira dashboard; `/dashboard` pode redirecionar a `/`.
3. Instalar shadcn/ui sobre Vite, alias e tokens Tailwind; sidebar e estados vazios
   reais, sem números fictícios aparentando dados financeiros reais.
4. Adicionar entidades e interfaces do domínio; validar broker com testes de contrato
   antes de espalhar endpoints externos pelo código.
5. Implementar feed e estratégias puras com fixtures; depois simulador histórico.
6. Implementar intenção/outbox/risco, executor paper e reconciliação de recuperação.
7. Conectar dados reais do backend ao dashboard e publicar na Railway quando solicitado.

## Definição de pronto

Critérios de aceite do SDD implementados; testes proporcionais ao impacto; schema
compatível; segredos fora do Git; estados de erro e loading presentes; log sem
credenciais; runbook atualizado. Para engine, testes de fault injection e prova
de que fechar o browser não interfere. CI verde e revisão dos limites da entrega.

## Rastreabilidade

| Requisito | Design | ADR | Verificação |
| --- | --- | --- | --- |
| RF-01–RF-04 | SDD-01 | 0001, 0003 | Auth + UI em `validacao.md` |
| RF-05 | SDD-02 a 04 | 0001, 0002, 0004 | Domínio, DB e integração |
| RF-06 | SDD-05 | 0005 | Deploy/restore/health |
| RP-01–RP-02 | SDD-02 | 0002, 0006 | Replay e integridade |
| RP-03, RP-05–RP-07 | SDD-03 | 0002, 0004, 0007 | Risco, idempotência e recuperação |
| RP-04 | SDD-04 | 0006, 0007 | Sem look-ahead e custos |
| RP-08–RP-09 | SDD-01, 03, 05 | 0003, 0005 | Autorização, auditoria e observabilidade |

## Dependências externas

Conta Railway e domínio/URL de callback; projeto OAuth Google; conta e credenciais
de corretora apenas após seleção; fonte histórica licenciada; orçamento; eventual
canal Telegram. Criar os projetos e publicar são etapas futuras, não entregas
concluídas por esta documentação. Não pedir valores de segredos em conversa.
