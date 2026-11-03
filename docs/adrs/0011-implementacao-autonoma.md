# ADR-0011 — execução autônoma e defaults do MVP

Data: 2026-11-03. Estado: aceito pela instrução posterior de implementar o roadmap
sem interrupções. Valores são defaults de engenharia para simulação, não uma
resposta inventada a Q7–Q9 nem recomendação de investimento.

## Decisão

Manter Brasil/EUA e capital exclusivamente fictício. Adotar watchlist PETR4,
VALE3, ITUB4, AAPL, NVDA, TSLA; long-only e sem leverage; candles fechados de cinco
minutos; fonte EUA IEX e importação CSV B3. Três famílias de estratégias, configs
imutáveis e desabilitadas por padrão. Adotar arquitetura modular com web e worker,
PostgreSQL sem extensões obrigatórias e shadcn/ui Radix.

Defaults de risco: 50 bps por operação; perda diária 100 bps; exposição 5000 bps;
três posições; custo simulado 5 bps por lado e slippage 5 bps. A política é validada
e editável com conta bloqueada; estratégias/contas começam sem executar.

## Consequências

Q7–Q9 deixam de bloquear a implementação; seu histórico permanece como não
respondido individualmente. Não reduzir Brasil a EUA. B3 tem calendário revisado
2026, rejeitando anos ainda não suportados; ampliar com fontes verificadas.
Sem horário B3 em tempo real ou corretora B3 externa. Não ligar dinheiro real.

Recuperação Google é CLI auditada e revoga sessões; não existe bypass público.
Deploy está autorizado como objetivo, mas exige as credenciais do proprietário.
Observação quantitativa P6 continua pendente de dados/período reais.

## Revisão

Alterar defaults por necessidade comprovada, com testes e documentação. Mudança
de mercado, capital real ou oferta para terceiros exige novo escopo.
