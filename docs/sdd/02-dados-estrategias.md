# SDD-02 — dados de mercado e estratégias

Estado: Brasil + EUA confirmados; contratos propostos. Corretoras, feeds, universo e timeframe pendentes.

## Entrada e contrato

Evento de mercado: `event_id`, `schema_version`, `source`, `instrument_id`,
`event_time`, `received_at`, sequência quando disponível e payload decimal.
Candle: timeframe, open/close time, OHLCV, encerrado, feed e revisão. Rejeitar
NaN/Inf, preços inválidos, OHLC incoerente, volume negativo e instrumento sem
metadados. Não substituir dado faltante por cotação zero ou preço inventado.

Adapter declara capacidades: histórico, streaming, atraso, calendário, limites e
ajustes. Persistir origem e revisão. Deduplicar, detectar gaps e reordenar dentro
de tolerância documentada; fora dela, reparar histórico e não emitir sinais
retroativos como se fossem oportunidades atuais. Bootstrap exige warmup completo.

## Estratégias puras e versionadas

Contrato lógico: `Evaluate(context, closed_candles, params, portfolio_snapshot)`
retorna `Signal` ou `NoSignal`, sem I/O externo. `Signal` tem origem, versão,
instante, instrumento, direção, razão e validade; ainda não é ordem aprovada.
Janela, warmup e validação pertencem à versão. Mesma entrada produz mesma saída.

Proposta inicial: cruzamento de médias; reversão à média via Z-score; rompimento
por faixa e volatilidade. Definir SMA vs EMA, janelas, confirmação, saída,
frequência e filtros após mercado. Desvio zero em Z-score resulta em no-signal.
Nenhuma estratégia usa candle futuro ou ainda aberto sem contrato específico.

Configuração é JSON/DSL declarativa com `schema_version` e allowlist de indicadores,
operadores e intervalos. Rejeitar campos desconhecidos e limites inválidos. Sem
`eval`, scripts enviados pelo usuário ou execução de texto de LLM. Alterações
criam versões imutáveis; ativação registra ator e versão da política de risco.

## Aceite e falhas

Testar warmup, repetição, gaps, DST, candle fora de ordem, revisão e reprocessamento.
Uma barra não produz duas intenções para a mesma execução/versionamento. Feed
vencido suspende novas exposições. A estratégia não chama broker diretamente.
Obter candles não concede licença para redistribuição a usuários de um futuro SaaS.
