# ADR-0014 — Railway atual: API de configuração e Dockerfile portátil

Estado: implementado em 2026-11-03, durante publicação autorizada pelo usuário.
Atualiza as instruções operacionais do ADR-0005 e ADR-0013.

## Contexto

A API Railway autenticada recusou `railwayConfigFile` com a mensagem de que
Config as Code (`railway.json`/`railway.toml`) foi descontinuado em favor de IaC.
O enum público `Builder` não inclui `DOCKERFILE`, embora o builder detecte um
Dockerfile automaticamente quando o caminho é configurado.

O primeiro build também recusou mounts sem ID e mounts do tipo secret. A
documentação atual de Docker exige IDs de cache literais no formato
`s/<service id>-<target path>` e não permite variáveis nesses IDs.
Fonte: https://docs.railway.com/builds/dockerfiles#cache-mounts.

## Decisão

- Configurar start command, migrations da web, readiness, uma réplica e
  `sleepApplication=false` via `serviceInstanceUpdate` da API atual.
- Versionar a operação em `scripts/railway-configure.sh`, invocada pelo script
  de deploy. Não passar `railwayConfigFile` nem o enum inexistente `DOCKERFILE`.
- Manter `railway.web.toml` e `railway.engine.toml` apenas como referências
  históricas dos valores; não orientar agentes a selecionar esses caminhos.
- Usar camadas Docker convencionais, sem mounts de cache/secret. Isso permite
  compartilhar o Dockerfile entre web, engine e CI sem IDs de serviço embutidos.
- Continuar com TLS validado e CA/tzdata na imagem final. Não enviar certificados
  ou segredos do executor Codex ao builder Railway.

## Alternativas e consequências

Dois Dockerfiles com IDs literais duplicariam a definição e a vinculariam aos
serviços atuais. IaC `.railway/railway.ts` é uma evolução possível, mas não é
necessário adotar um runtime de autoria adicional para esta configuração pequena.
O custo de remover mounts é menor reutilização de cache fino; as camadas de
dependências continuam reutilizáveis quando seus manifests não mudam.

Reavaliar com necessidade de ambientes adicionais, drift ou builds frequentes.
Uma nova mudança da API deve ser inspecionada no schema antes de alterar serviços.
