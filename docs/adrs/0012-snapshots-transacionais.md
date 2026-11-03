# ADR-0012 — snapshots financeiros transacionais e artefatos imutáveis

Data: 2026-11-03. Estado: implementado no escopo pessoal delegado.

## Contexto

O modelo lógico anterior enumerava várias tabelas. No MVP de conta única, a
invariante fundamental é persistir a reserva e intenção atomicamente e preservar
fills idempotentes. Separar cada entidade em tabela acrescentaria coordenação.

## Decisão

`trading_accounts` guarda snapshot JSON versionado por conta com ordens/intents,
reservas, posições, fills, decisões, políticas, heartbeat, cursores e gerações.
Toda mutação usa transação, lock de linha PostgreSQL e versão esperada. A outbox
é a lista durável de ordens aprovadas no próprio snapshot; envio externo ocorre
depois do commit. Decimais shopspring/decimal são serializados como strings,
sem floats para caixa, limites ou fills. JSON fica em coluna TEXT portátil para
SQLite de teste; não alegar NUMERIC normalizado onde não foi usado.

`datasets` e `strategies` guardam configurações imutáveis; `backtests` mantém
job/estado/resumo e `artifacts` preserva curva/fills completos fora do polling.
`audit_events` registra ações e alterações de política antes/depois na mesma
transação, incluindo rejeições de risco. Ator é owner/engine/operator.

## Tradeoffs

Simplifica invariantes e replay, mas o snapshot cresce com o histórico e é
reescrito nas mutações. Este é desenho de terminal pessoal, não HFT/SaaS.
Não truncar ledger para ganhar espaço. Antes de volumes elevados, migrar fills,
ordens e decisões para tabelas normalizadas com reconciliação e comparação de
saldos. Não foi feito benchmark para milhões de eventos.

## Recuperação e ambiguidade

Timeout de envio mantém intenção unknown; não há retry cego. Fills cumulativos
são reconciliados pela diferença de quantidade e notional, com deduplicação.
Snapshot em PostgreSQL não oferece transação atômica com a corretora.

Novo replay B3 exige bloqueio e ausência de exposição/ordens; o snapshot anterior
é arquivado integralmente em `artifacts`, na mesma transação do reset, com link
na auditoria. Nenhum fill anterior é apagado.
