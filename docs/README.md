# Scooby — documentação do produto e da arquitetura

O [ROADMAP.md da raiz](../ROADMAP.md) é o documento principal do projeto.
Leia-o primeiro para onboarding, estado real e próxima ação; este índice reúne
os documentos de referência.

Data-base: 2026-11-03. Idioma dos documentos: português.

**Estado atual: implementação paper entregue, publicação/configuração externa pendentes.**
O [ROADMAP](../ROADMAP.md) e os ADRs 0011–0013 descrevem código e defaults atuais.
Os desenhos abaixo preservam o planejamento original; onde houver diferença,
consultar os ADRs de implementação e [checklist de deploy](deploy-checklist.md).

**Histórico do planejamento:** Esta documentação não representa funcionalidades já implementadas. Os requisitos expressos pelo usuário estão confirmados; recomendações e ADRs propostos precisam das respostas ao grilling. Não há autorização de ativação de operações reais implícita neste planejamento.

## Navegação

| Documento | Finalidade |
| --- | --- |
| [Produto](produto.md) | Objetivo, escopo e requisitos rastreáveis |
| [Brainstorming](brainstorming.md) | Alternativas, tradeoffs e hipóteses |
| [Grilling](grilling.md) | Árvore de decisões e rodadas de perguntas |
| [Arquitetura](arquitetura.md) | Serviços, fluxo de eventos e limites |
| [Modelo de dados](modelo-de-dados.md) | Entidades, integridade e precisão |
| [Planejamento](planejamento.md) | Fases, dependências e critérios de conclusão |
| [Operação na Railway](railway.md) | Deploy, variáveis, recuperação e observabilidade |
| [Validação](validacao.md) | Testes, experimentos quantitativos e liberação |
| [Fontes](fontes.md) | Documentação primária e limites das verificações |
| [Deploy/configurações](deploy-checklist.md) | Variáveis, publicação e recuperação |
| [Pesquisa web](web-research.nd) | Achados, fontes oficiais, limitações e impactos no projeto |
| [SDDs](sdd/README.md) | Especificações de comportamento e desenho técnico |
| [ADRs](adrs/README.md) | Decisões arquiteturais e respectivos estados |

## Escopo confirmado pelo proprietário

Uso pessoal, mercados Brasil/B3 e EUA, backtesting e paper trading no MVP.
Alpaca Paper nos EUA, simulador interno B3 e uma conta Google autorizada estão
confirmados. Fontes de dados, instrumentos, frequência, limites e recuperação da
identidade continuam em refinamento. Rodada 3 em `grilling.md`.

## Como ler os estados

- **Confirmado:** requisito explicitamente solicitado nesta conversa.
- **Proposto:** alternativa recomendada, ainda não aceita pelo usuário.
- **Pendente:** depende de uma resposta ou investigação.
- **Implementado:** evidência no código, distinta do estado de uma decisão.

## Inventário real da base

Repositório: `AlexGalhardo/Scooby`. Base inspecionada: commit `9e21d88`.
Vukem 0.37.0; Go 1.27.3; Inertia React 3.9.1; React 19.3.1;
Tailwind 4.3.3; PostgreSQL com driver Vukem 0.6.1; Vite e TypeScript.
O inventário acima identifica a base original. A evolução implementou Google-only,
sidebar shadcn/ui, dashboard, dados, estratégias, backtest, adapters paper e worker.
`/posts` não é mais publicado. Estado detalhado e testes no ROADMAP.

## Processo utilizado

A skill `grilling` foi encontrada em `/tmp/vukem/.claude/skills/grilling/SKILL.md`.
Ela orienta entrevistas em rodadas, com dependências explícitas entre decisões.
Não foram encontradas skills instaladas com os nomes `planejamento`,
`brainstorming` ou `grill-me`; os documentos com esses temas são trabalho de
planejamento, sem alegar execução de skills ausentes. A cópia em `/tmp` não faz
parte do projeto e pode desaparecer; este registro preserva a procedência.

O usuário autorizou documentar agora. A implementação do produto não começa
nesta etapa. Depois de cada resposta, atualizar `grilling.md`, requisitos,
SDDs afetados e ADRs. Nunca promover uma recomendação a decisão silenciosamente.
