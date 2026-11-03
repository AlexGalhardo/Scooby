# SDD-06 — adapters confirmados e modalidades de simulação

> Implementação atual e desvios explícitos em [ROADMAP](../../ROADMAP.md)
> e [ADRs 0011–0013](../adrs/README.md). Aceite técnico local é distinto
> de OAuth/broker reais, publicação Railway e observação financeira prolongada.

Estado: Alpaca Paper EUA e simulação interna B3 confirmados; dados em Q8.

## Adapter de corretora

Contrato proposto: obter capacidades/instrumentos, conta, ordens abertas,
consulta por client ID, enviar/cancelar ordem, execuções por cursor e posições.
Respostas distinguem rejeição definitiva, limitação temporária, timeout e estado
indeterminado. `Capabilities` explicita mercado, mode, tipos de ordem, time-in-force,
quantidade mínima/step, consulta por ID e comportamento idempotente.

Alpaca Paper usa `https://paper-api.alpaca.markets` com credenciais paper. Host de
dados é configuração separada; não inferir que feed também está no host paper.
No boot conferir modo/conta e bloquear hosts live no MVP. Respeitar rate limits,
backoff com jitter para leituras, e política de reconciliação para mutações.
Testes de contrato não usam a conta pessoal no CI.

## Adapter B3 interno

Simular aceitação, fills, parciais, rejeições, cancelamento e latência com relógio
controlado. Não possui conexão de execução à bolsa. Aceitar apenas instrumentos
com metadados de tick, lote e sessão; os parâmetros finais dependem de Q7/Q9.
Regra de fill/custos versionada, sem liquidez ilimitada implícita. Dados sintéticos
servem fixtures e demonstração com etiqueta, não validação da estratégia.

## Importação de histórico — proposta em Q8

CSV documentará símbolo/mercado/feed, timezone/offset explícito, timeframe,
open_time, OHLCV, moeda e revisão. Importar com manifesto de origem/licença,
intervalo, hash do arquivo e política de ajustes. Rejeitar duplicatas conflitantes,
horários ambíguos e intervalos inválidos; informar linhas inválidas sem importar
silenciosamente só as convenientes. Limitar tamanho e tratar nomes de arquivo
como entrada não confiável. A forma física do armazenamento do arquivo original
na Railway será definida com retenção; banco guarda manifesto/hash e candles.

## Estados na interface e aceite

Mostrar `Alpaca Paper`, `Simulação B3 — replay` ou `Simulação B3 — tempo real`
conforme relógio/fonte efetivamente implementados. Dataset ausente → vazio/ação de
importar; API desconectada → degradado; preço vencido → bloqueio de entradas.
Nenhum saldo ou fill pode migrar entre contas/modos. Testar cancelamento parcial,
deduplicação, fonte ausente, rate limit e crash antes/depois do envio.
