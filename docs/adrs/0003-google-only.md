# ADR-0003 — autenticação exclusiva Google

Data: 2026-11-03. Estado: **Aceito por requisito; uma conta autorizada confirmada**.

## Contexto

O usuário exige `/login` com apenas Google e redirect de visitante do dashboard para essa página.

## Decisão

Usar OAuth/OIDC Google no servidor com suporte Vukem; sessões locais. Remover rotas locais de senha/cadastro/reset do scaffold. Login bem-sucedido termina em `/`.

## Alternativas

Senhas locais, múltiplos provedores e login diretamente na corretora. Não correspondem ao fluxo solicitado.

## Consequências

Depende de projeto OAuth, callback e armazenamento compartilhado de states/sessões. Autenticação não substitui autorização; uso pessoal confirmado, Q6 confirmou uma conta, autorizada por email verificado no bootstrap e depois vinculada ao subject Google.

## Gatilho de revisão

Novo requisito de identidade ou recuperação. Não adicionar provedor sem decisão registrada.
