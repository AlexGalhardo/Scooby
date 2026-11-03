# ADR-0010 — Alpaca Paper e simulador interno B3

Data: 2026-11-03. Estado: **Aceito nas respostas Q4 e Q5; implementação pendente**.

## Contexto

O MVP cobre Brasil e EUA sem dinheiro real. Não há integração de ordens B3
comprovada ou corretora brasileira selecionada.

## Decisão

Usar Alpaca Paper como primeira integração americana. B3 começa com backtest e
simulador interno; integração paper externa só após validar API/contrato da
corretora. Adapter de dados separado do de execução. Feed IEX e importação CSV
B3 são propostas de Q8, não parte automaticamente aceita deste ADR.

## Alternativas

IBKR, exigir corretora B3 imediatamente ou suportar só EUA. As duas primeiras não
foram escolhidas; a última conflita com os dois mercados confirmados.

## Consequências

Dashboard distingue ordens paper da corretora e fills simulados internos. São
contas/ledgers separados. Sem provedor de dados, simulador interno não recebe
cotações reais por mágica: histórico e tempo real precisam de fonte definida.
Regras de fill, custos e dados revisados devem ficar em cada relatório.

## Gatilho de revisão

API B3 verificada, limitação do plano Alpaca ou pedido de outro broker. Implementar
novo adapter com contratos, não trocar URL paper para live.
