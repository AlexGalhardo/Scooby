# ADR-0006 — PostgreSQL inicial sem extensões obrigatórias

Data: 2026-11-03. Estado: **Proposto**.

## Contexto

A conversa cita TimescaleDB, Redis e NATS, mas volume, mercado e orçamento não estão definidos.

## Decisão

Propor PostgreSQL convencional e outbox primeiro; adicionar TimescaleDB, Redis ou NATS só por necessidade demonstrada e suporte validado. Manter contratos versionados de evento.

## Alternativas

Operar todos desde o dia um; depender de extensão não verificada; eventos apenas em memória. Acrescentam custo ou perda de durabilidade.

## Consequências

Polling inicial adiciona latência e exige limites; desempenho será medido. Dados de tick/book ou alto fan-out podem alterar a decisão.

## Gatilho de revisão

Medição de throughput, retenção, latência ou consumidores justificar serviço/extensão e orçamento aprovado.
