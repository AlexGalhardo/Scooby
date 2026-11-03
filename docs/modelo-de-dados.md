# Modelo de dados proposto

> Atualização de implementação: consulte o [ROADMAP](../ROADMAP.md),
> [ADR-0011](adrs/0011-implementacao-autonoma.md),
> [ADR-0012](adrs/0012-snapshots-transacionais.md) e
> [ADR-0013](adrs/0013-workers-polling-e-deploy.md) para o estado efetivo.
> O texto abaixo preserva requisitos/desenhos originais, não uma declaração de
> que todos os detalhes de infraestrutura ou validação externa já ocorreram.

Estado: desenho lógico, sem migrations implementadas. Público e mercado ainda
podem alterar cardinalidades e precisão.

## Entidades

| Tabela | Campos e restrições principais |
| --- | --- |
| `users` | id, nome, email verificado, status, timestamps |
| `social_accounts` | user_id, provider, subject; UNIQUE(provider, subject) |
| `sessions` | sessão revogável, expiração, usuário; sem token Google desnecessário |
| `trading_accounts` | owner_id, broker, external_id, mode, currency, status |
| `instruments` | mercado, símbolo, moeda, tick_size, qty_step, timezone, status |
| `candles` | instrument_id, feed, timeframe, open_time, OHLCV, closed, revision |
| `strategy_versions` | strategy_id, versão, schema_version, parâmetros JSONB, hash |
| `strategy_runs` | account_id, versão, mode, status, início/fim e dataset |
| `signals` | run_id, candle/evento, direção, razão; dedupe_key UNIQUE |
| `risk_policy_versions` | conta, versão, moeda, limites e validade |
| `risk_decisions` | sinal, versão, snapshot, motivo, aprovado/rejeitado |
| `risk_reservations` | conta, intenção, exposição reservada, status |
| `order_intents` | conta, sinal, idempotency_key, client_order_id, estado |
| `broker_orders` | conta, broker_order_id UNIQUE por conta, intenção, status |
| `fills` | conta, broker_execution_id UNIQUE por conta, qty, price, fees |
| `ledger_entries` | lançamentos de caixa/posição, fill e reversão referenciados |
| `position_snapshots` | conta/instrumento, qty, custo e marcação; projeção recuperável |
| `outbox_events` | tipo, versão, payload, tentativas, available_at, claim e geração |
| `engine_controls` | conta, halted, motivo, generation, heartbeat e leader |
| `reconciliation_runs` | conta, resultado, divergências, cursor e horário |
| `backtest_runs` | código, dataset/hash, parâmetros, seeds, custos e métricas |
| `audit_events` | ator, conta, ação, versão anterior/nova, request_id e horário |

## Precisão e unidades

Proposta: PostgreSQL `NUMERIC` e decimal explícito em Go; strings decimais nos
contratos JSON e conversão de exibição no React. Escalas finais dependem dos
instrumentos; não assumir duas casas para preço, quantidade ou cripto. Arredondar
quantidade para baixo conforme lote/step ao abrir risco, preço conforme tick e
lado; validar mínimos e máximos do broker. Não usar floats para ledger ou limites.
Indicadores estatísticos podem usar float com tolerâncias e entradas finitas.

Toda quantia carrega moeda. Não somar USD e BRL sem taxa, instante e fonte de FX.
Sem fonte confiável, mostrar valores separados. Armazenar UTC (`timestamptz`),
conservar timezone de mercado e usar calendário oficial para feriados e DST.

## Integridade e concorrência

- FKs preservam conta/execução/versão; IDs externos são únicos no escopo correto.
- Candles: UNIQUE(instrument_id, feed, timeframe, open_time, revision); uma revisão
  corrente explícita evita misturar revisões no mesmo backtest.
- Reservas de risco e intenção são criadas atomicamente com outbox. Incluir ordens
  parcialmente executadas e em estado desconhecido no cálculo conservador.
- Eventos de fill deduplicados antes de alterar ledger. Correções são lançamentos
  de ajuste/reversão, nunca apagamento silencioso de evidência financeira.
- Contas real e paper, datasets sintéticos e históricos usam namespaces explícitos.
- Índices por conta+status+tempo em ordens e conta+instrumento em posições.
- Retenção/particionamento de candles só após orçamento e volume; não descartar
  dados que sustentem resultados sem registrar indisponibilidade do dataset.

## Risco e sizing

Para uma compra sem alavancagem: `budget = equity * risk_fraction`;
`risk_per_unit = entry - stop + estimated_costs_per_unit + slippage_buffer`;
`qty = floor_to_step(min(budget/risk_per_unit, cash_limit/entry, exposure_limit/entry))`.
Rejeitar stop >= entrada, custos negativos, denominador não positivo, valores não
finitos ou quantidade abaixo do mínimo. A reserva usa cenário conservador de
execução, não lucro esperado. Stops não limitam perda máxima garantida.

P&L realizado vem dos fills, política de custo e taxas; não realizado exige preço
marcado recente e fonte. Resultado antes de custos, após custos e estimativa
tributária são campos distintos; apuração fiscal não está definida no MVP.

## Ajuste confirmado: dois mercados, uso pessoal

O modelo deve distinguir Brasil e EUA e preservar contas paper separadas. Um
proprietário pode possuir várias `trading_accounts`; `owner_id` serve autorização,
sem introduzir organizações, billing ou tenancy SaaS no MVP. Unicidade de
instrumento usa mercado/venue/símbolo, não símbolo isolado. Separar BRL e USD no
ledger; política de FX e limites agregados ainda dependem de decisão. Dados
sintéticos precisam de flag e fonte visíveis, inclusive no simulador interno.
