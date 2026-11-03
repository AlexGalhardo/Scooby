# ADR-0001 — stack e dashboard

Data: 2026-11-03. Estado: **Aceito por requisito; implementação parcial**.

## Contexto

O repositório já usa Vukem/React/Inertia/Tailwind/PostgreSQL. O usuário exige dashboard em `/` com sidebar shadcn/ui.

## Decisão

Manter a stack; adaptar o starter ao dashboard autenticado e instalar componentes shadcn/ui no frontend Vite. `/` é a rota canônica, não uma landing page pública.

## Alternativas

SPA com API separada; outro framework; `/dashboard` como destino principal. Não atendem melhor aos requisitos já fixados.

## Consequências

Compartilhar tipos/props entre Go e React; preservar integração Inertia e CSP. shadcn/ui exige configuração e componentes que ainda não existem no código.

## Gatilho de revisão

Mudança explícita de requisitos ou incompatibilidade reproduzida na integração.
