# Visão do produto e requisitos

Estado: requisitos confirmados + escopo recomendado em discussão.

## Objetivo

Criar o Scooby como plataforma de negociação algorítmica: pesquisar estratégias,
simular seus resultados, acompanhar mercado e executar regras com controles de
risco. Lucratividade não é requisito garantível. A conversa fornecida é contexto
e catálogo de possibilidades; seus exemplos não confirmam mercado, corretora,
valores de risco, orçamento ou finalidade comercial.

## Requisitos confirmados

| ID | Requisito | Especificação |
| --- | --- | --- |
| RF-01 | `/` é o dashboard principal autenticado | SDD-01 |
| RF-02 | Visitante sem sessão é redirecionado para `/login` | SDD-01 |
| RF-03 | `/login` oferece somente o botão de login Google | SDD-01 |
| RF-04 | Dashboard padrão com sidebar em shadcn/ui | SDD-01 |
| RF-05 | Manter Go/Vukem, React/Inertia, Tailwind e PostgreSQL | ADR-0001 |
| RF-06 | Deploy na Railway | SDD-05, ADR-0005 |
| RF-07 | Planejamento, SDDs e ADRs versionados em `docs/` | Este conjunto |
| RF-08 | Refinar decisões por grilling | `grilling.md` |

Google autentica a pessoa no Scooby; não concede acesso à conta da corretora.
O redirect automático é do dashboard protegido para `/login`; o navegador só
sai para o Google quando a pessoa clica no botão.

## Decisões confirmadas na rodada 1

| ID | Decisão | Origem |
| --- | --- | --- |
| RF-09 | Terminal pessoal no MVP | Resposta Q1 |
| RF-10 | Cobrir Brasil/B3 e EUA | Resposta Q2 |
| RF-11 | MVP com backtesting e paper trading; sem capital real | Resposta Q3 |
| RF-12 | Alpaca Paper como primeira integração EUA | Resposta Q4 |
| RF-13 | B3 começa com simulador interno | Resposta Q5 |
| RF-14 | Somente uma conta Google autorizada | Resposta Q6 |

Ambos os mercados fazem parte do escopo; uma entrega incremental por mercado
não remove a B3 do MVP. Alpaca Paper e simulação interna B3 estão definidos; datasets, orçamento e
modalidade temporal da simulação B3 continuam em refinamento. Cada mercado terá calendário, moeda, precisão
e adapter próprios. Contas e relatórios permanecem segregados por BRL e USD até
existir política de consolidação de FX. Uso pessoal não significa cadastro aberto.

## Escopo funcional proposto

| ID | Capacidade | Aceite resumido |
| --- | --- | --- |
| RP-01 | Séries de mercado com procedência | Símbolo, feed, horário, atraso e gaps visíveis |
| RP-02 | Estratégias versionadas | Parâmetros validados, versão imutável, ativação auditada |
| RP-03 | Motor de risco central | Toda ordem validada, incluindo ordens manuais futuras |
| RP-04 | Backtesting reproduzível | Dataset, código, configuração e custos identificados |
| RP-05 | Paper trading | Execução independente do navegador e reconciliação |
| RP-06 | Ordens, fills e posições | Histórico persistente e estados ambíguos explícitos |
| RP-07 | Emergência | Bloqueia novas exposições e permite acompanhar cancelamentos |
| RP-08 | Dashboard operacional | P&L, exposição, drawdown, ordens e saúde dos serviços |
| RP-09 | Auditoria | Quem alterou o quê, versão, horário e correlação |

## MVP recomendado, ainda não aprovado

Brasil e EUA confirmados, contas de simulação segregadas por mercado; candles em vez de HFT/book completo,
três estratégias (médias, reversão e rompimento), capital fictício, sem alavancagem
ou venda a descoberto inicialmente. Limites concretos serão escolhidos depois
do mercado e instrumento. Pesquisa em Python pode ser adicionada quando houver
necessidade demonstrada; não exige segundo runtime no primeiro deploy.

SaaS, sinais pagos, derivativos, arbitragem estatística, ML, execução real e
apuração fiscal ficam fora do MVP confirmado. Integrações múltiplas são previstas
para Brasil e EUA, com Alpaca Paper confirmada e corretora externa B3 adiada até validação de API. Podem voltar em uma fase futura mediante decisão de produto.

## Qualidade e limites

Dinheiro e quantidades não usam ponto flutuante binário na contabilidade. Toda
mutação exige autorização no servidor. Dados obsoletos não geram novas exposições.
Sem estado persistente ou reconciliação confiável, a engine não retoma operações.
Fechar o navegador ou terminar a sessão não interrompe o worker.

Ajustar metas mensuráveis de latência, disponibilidade, retenção e capacidade na
rodada de operação. Não prometer baixa latência de bolsa em infraestrutura de
aplicação geral. Distribuição de dados, elegibilidade da corretora, tributação e
requisitos de oferecer o serviço a terceiros precisam de análise própria após
as respostas de público e mercado; este documento não faz conclusão jurídica.
