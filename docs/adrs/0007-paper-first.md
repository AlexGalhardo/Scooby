# ADR-0007 — validação por simulação antes de live

Data: 2026-11-03. Estado: **Aceito para paper-first pela resposta a Q3; detalhes técnicos propostos**.

## Contexto

Execução automatizada precisa de evidência de correção e recuperação. O usuário confirmou backtesting e paper trading no MVP em Q3.

## Decisão

MVP backtest + paper, sem adapter live habilitado. Usar estratégias e risco determinísticos; LLM, se futura, só produz configuração revisável.

## Alternativas

Live desde o início ou LLM com acesso direto a ordens. O primeiro amplia critérios de escopo; o segundo não preserva o motor determinístico proposto.

## Consequências

Permite testar contratos sem capital real, mas paper não prova lucro nem fidelidade de liquidez. Broker paper ainda é decisão posterior ao mercado.

## Gatilho de revisão

Futura mudança explícita de escopo, com limites, evidências e autorização operacional própria.
