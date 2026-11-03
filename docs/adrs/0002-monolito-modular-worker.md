# ADR-0002 — domínio modular e worker contínuo

Data: 2026-11-03. Estado: **Proposto**.

## Contexto

Execução de estratégias precisa sobreviver ao fechamento do dashboard; microsserviços precoces elevam custo de consistência.

## Decisão

Propor código de domínio Go compartilhado e processos separados web/engine. Worker possui liderança por conta, geração e reconciliação no boot.

## Alternativas

Engine dentro de requests/browser; único processo acoplado; vários microsserviços. O primeiro não atende continuidade; os demais mudam isolamento/custo.

## Consequências

Mesmo repositório e banco simplificam transações; deploy de dois processos exige contratos compatíveis e fencing. Lock não cancela requisições externas em voo.

## Gatilho de revisão

Volume, isolamento de falha ou equipe justificarem extração de módulos em serviços.
