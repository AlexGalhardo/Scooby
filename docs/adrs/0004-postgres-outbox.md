# ADR-0004 — persistência financeira e outbox

Data: 2026-11-03. Estado: **Proposto; PostgreSQL confirmado**.

## Contexto

Ordens e reservas precisam ser duráveis e consistentes; rede externa não participa da transação do banco.

## Decisão

Propor PostgreSQL como registro de intenções, fills, ledger, reservas e outbox transacional. NUMERIC/decimal para valores; unicidade de sinais e IDs de execução.

## Alternativas

Redis como banco definitivo; mensageria sem outbox; floats na contabilidade. Não oferecem as invariantes necessárias ao desenho proposto.

## Consequências

Exige locks por conta, índices, workers idempotentes e limpeza/monitoramento de outbox. Não entrega exactly-once na corretora; timeout exige reconciliação.

## Gatilho de revisão

Contenção ou volume medidos exigirem outro transporte, preservando registro e invariantes.
