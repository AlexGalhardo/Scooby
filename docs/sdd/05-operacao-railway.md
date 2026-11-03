# SDD-05 — entrega e operação na Railway

Estado: Railway confirmado; desenho e metas propostos.

## Contrato de publicação

Imagem versionada, migrations compatíveis, segredos por serviço e separação de
web/worker. Engine deve permanecer ativa sem tráfego HTTP. Banco compartilhado
em rede privada. Zero alteração em endpoints paper/live por decisão automática
de deploy. Worker entra reconciliando, não em modo de execução imediato.

Artefatos futuros: configuração Railway por serviço, build dos dois binários,
readiness, comando de migration com lock, métricas, alertas e runbooks. Não há
serviço engine nem configuração Railway implementados no starter atual.

## Cenários de aceite

- Deploy com migration falha não promove nova versão web/engine.
- Sobreposição de workers não permite duplicação de intenção/envio por novo líder.
- SIGTERM preserva checkpoint e não perde preenchimento recebido durante shutdown.
- Restart mantém kill switch, perda diária e ordens desconhecidas.
- Falta de sessão de navegador não interfere no heartbeat.
- Readiness falha quando dependências essenciais impedem servir/operar; liveness
  é distinta e não divulga dados privados.
- Restore testado em conta sem execução; reconciliar antes de retomar.
- Alertas de feed vencido, worker parado e ordem ambígua funcionam sem dashboard.

## Pendências operacionais

RPO, RTO, SLO, região, orçamento, retention, alerta Telegram ou outro canal,
limite de carga e responsável por incidentes dependem da rodada operacional.
Detalhes e variáveis em `../railway.md`; não copiar segredos para SDDs.
