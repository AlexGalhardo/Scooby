# ADR-0005 — hospedagem na Railway

Data: 2026-11-03. Estado: **Aceito como destino; topologia proposta**.

## Contexto

O usuário definiu Railway para deploy; a execução precisa de serviço contínuo.

## Decisão

Hospedar na Railway. Propor serviços web e engine separados, PostgreSQL privado, migrations em pre-deploy e engine sem hibernação.

## Alternativas

VPS, Kubernetes ou execução em funções efêmeras. Alternativas não são o destino pedido; podem ser revisitadas se limitações medidas exigirem.

## Consequências

Configurar saúde, segredos, recursos, deploy concorrente e backup. Railway não garante sozinho exclusividade de executor nem reconciliação.

## Gatilho de revisão

Limitação de região, conectividade, custo, disponibilidade ou requisitos específicos do broker.
