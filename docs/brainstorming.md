# Brainstorming — alternativas e hipóteses

Estado: exploração de alternativas. Rodada 1 confirmou uso pessoal, Brasil + EUA e backtest/paper; alternativas incompatíveis permanecem apenas como histórico de análise.

## Alternativas de produto

| Caminho | Benefício | Custo ou dependência | Recomendação |
| --- | --- | --- | --- |
| Terminal pessoal | Uma identidade e conta simplificam operação | Allowlist e recuperação do proprietário | Começar aqui |
| SaaS | Vários usuários e contas | Isolamento, autorização, cobrança, suporte e análise regulatória | Só se Q1 exigir |
| Plataforma de pesquisa | Aprende sem conectar execução | Difere do resultado em mercado | Primeiro incremento |
| Terminal com paper broker | Testa execução e recuperação | Credenciais e plano de dados | Próximo incremento |

## Mercado e corretora

| Alternativa | Questão técnica a comprovar |
| --- | --- |
| EUA / Alpaca | Elegibilidade, feed contratado, ajustes e condições paper/live |
| EUA / Interactive Brokers | Sessões, permissões e operação do gateway escolhido |
| B3 | Corretora efetivamente oferece execução via API? Contrato e histórico disponíveis? |
| Cripto | Bolsa, jurisdição, custódia, precisão por par e funcionamento 24/7 |
| Vários mercados | Calendários, moedas, contratos, feeds e adapters distintos |

Alpaca é candidata se o mercado inicial for EUA; não foi escolhida. Dados de
mercado e API de ordens são contratos distintos. Não tratar PETR4 como ativo
operável na Alpaca nem presumir que todo feed gratuito representa o mercado todo.

## Alternativas de arquitetura

- **Monólito modular + worker separado:** compartilha domínio Go, mas processos
  web e engine têm ciclos de vida e credenciais distintos. Recomendado no MVP.
- **Microsserviços por motor:** isolamento maior, mais rede, deploy e consistência
  distribuída. Adiar até que uma carga ou equipe justifique a separação.
- **Eventos PostgreSQL + outbox:** transação local e recuperação simples, com
  polling controlado. Escolha proposta para começar.
- **NATS JetStream:** útil com vários consumidores e replay; implica operar o
  broker. Entrega de mensagens não elimina idempotência na corretora.
- **Redis:** pode atender cache/sessões, mas não ser o registro definitivo de ordens.
- **TimescaleDB:** avaliar extensão e suporte efetivo na Railway antes de depender
  dela. PostgreSQL convencional com índices/partições é a opção inicial proposta.

## Hipóteses quantitativas

Médias móveis: testar cruzamentos com regras explícitas de entrada/saída.
Reversão: Z-score com janela mínima e guarda para desvio zero; limiares não são
recomendação de investimento. Rompimento: faixa anterior e volatilidade sem
usar máxima futura do próprio candle. Cada estratégia precisa de regras de
sizing, saída, sessão e dados ausentes. Um bom resultado in-sample não valida a
estratégia. Não escolher parâmetros com base no conjunto reservado de avaliação.

## O que ainda não adotar

LLM não assina nem envia ordens. Se adicionada, produz rascunho de DSL validada e
revisada; nunca executa código arbitrário. Não prometer detectar topo/fundo,
"exactly once" em rede externa, stop a preço garantido ou retorno financeiro.

## Perguntas para próximos experimentos

O feed cobre os ativos desejados? O orçamento comporta histórico e licenças?
Qual diferença entre backtest, simulação interna e paper da corretora? Em quais
cenários uma estratégia deve ser desativada? Quais métricas realmente ajudam o
proprietário a decidir, em vez de apenas aumentar a aparência do dashboard?
