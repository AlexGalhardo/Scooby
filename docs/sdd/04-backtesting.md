# SDD-04 — backtesting e pesquisa

> Implementação atual e desvios explícitos em [ROADMAP](../../ROADMAP.md)
> e [ADRs 0011–0013](../adrs/README.md). Aceite técnico local é distinto
> de OAuth/broker reais, publicação Railway e observação financeira prolongada.

Estado: proposto. Dataset, frequência, benchmark e custos pendentes.

## Contrato de uma execução

Entrada: versão de estratégia, instrumentos, intervalo e calendário, dataset
imutável/hash, seed quando aplicável, capital e moeda, modelo de fill, taxas,
spread, slippage, latência e política de risco. Guardar commit/hash do código e
versão do runtime. Saída: sinais, decisões de risco, ordens/fills simulados,
curva de equity, métricas e relatório de limitações.

Utilizar o mesmo contrato de estratégia/risco do worker com relógio e broker
simulados. Replay roda isolado da execução e não possui credenciais live. Job
assíncrono durável, estados queued/running/completed/failed/canceled, orçamento de
CPU/memória e possibilidade de cancelar. Não bloquear request HTTP enquanto testa.

## Simulação

No padrão proposto, candle encerrado gera sinal executável só no próximo evento
elegível. Fills respeitam tipo de ordem, tick, lote, custos e liquidez modelada.
OHLC não revela trajetória intrabar: se stop e alvo cabem na mesma barra, usar
regra conservadora explícita ou dados mais finos; nunca escolher melhor resultado
silenciosamente. Gaps podem executar além do stop, com hipótese registrada.

Custos e impostos não equivalem a uma constante universal. Parametrizar custos
operacionais; estimativa fiscal separada, se implementada, com regras revisadas.
Comparar cenários de slippage/liquidez e reportar sensibilidade. Falha ou ausência
de histórico faz a execução falhar ou marca intervalo incompleto explicitamente.

## Relatório e aceite

Reproduzir mesma curva/fills dentro de tolerância documentada com mesmas entradas.
Identificar antes/depois de custos, benchmark, drawdown, tamanho da amostra e
partições temporais. Sem tuning no teste final. Comparar com fixtures manuais
simples e testar zero trades, perdas consecutivas e patrimônio insuficiente.
Paper interno e paper de corretora são resultados separados, com suas hipóteses.
Go é runtime recomendado inicialmente; Python opcional troca artefatos versionados,
sem alterar estratégia de produção sem novo teste de equivalência.
