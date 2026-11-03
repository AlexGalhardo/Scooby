# ADR-0013 — polling, liderança e deploy com standby

Data: 2026-11-03. Estado: implementado no escopo delegado.

## Decisão

Dashboard faz polling autenticado de snapshots a cada 10 segundos. Worker
consulta a corretora/feed e faz reconciliação por REST, em loop de 15 segundos;
WebSocket/SSE não são necessários para este MVP de candles 5m. Processamento
histórico ocorre em goroutine separada da execução, com timeout/cancelamento.

Worker mantém advisory lock PostgreSQL em conexão dedicada. Outro processo
permanece em standby até a liberação; não executa jobs ou ordens. Ao assumir,
incrementa leader_epoch e geração, recupera envios interrompidos como unknown e
reconcilia antes de novas entradas. Cada mutação operacional verifica epoch.

## Consequências

Readiness da engine verifica heartbeat compartilhado: permite promoção do novo
container em standby enquanto o antigo está saudável. Depois da parada do antigo,
a liderança troca sem dois executores ativos. Se não houver liderança saudável,
a readiness falha. Um request externo já em voo continua sujeito a reconciliação.

Políticas de encerramento e stops da engine são observadas, não ordens stop
nativas garantidas na corretora. Indisponibilidade pode impedir saída; UI e
runbook explicam. Somente paper está implementado.

## Revisão

Extrair o backtest para serviço separado se a carga comprometer heartbeat.
Adicionar streams se a latência observada justificar; manter a reconciliação REST.
Testar takeover, cancelamento e restarts após qualquer mudança de liderança.
