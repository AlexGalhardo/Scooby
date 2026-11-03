# Fontes e verificações

Data de consulta: 2026-11-03. Preferência por documentação primária. URLs apoiam
capacidades técnicas; escolhas de arquitetura deste projeto são propostas nossas.

| Fonte | Uso |
| --- | --- |
| [Vukem — introdução](https://vukem.dev/docs/introduction) | Framework existente |
| [Vukem — social login](https://vukem.dev/docs/security/social-login) | Google, OIDC, PKCE, state, scaffold e variáveis |
| [Vukem — Inertia](https://vukem.dev/docs/frontend/inertia) | Páginas React e props no backend |
| [shadcn/ui — Vite](https://ui.shadcn.com/docs/installation/vite) | Integração prevista com o frontend atual |
| [shadcn/ui — Sidebar](https://ui.shadcn.com/docs/components/base/sidebar) | Componente requerido |
| [Google — OpenID Connect](https://developers.google.com/identity/openid-connect/openid-connect) | Identidade e validação do login |
| [Railway — pre-deploy](https://docs.railway.com/deployments/pre-deploy-command) | Migrations antes de servir versão |
| [Railway — healthchecks](https://docs.railway.com/deployments/healthchecks) | Verificação na publicação |
| [Railway — scaling](https://docs.railway.com/deployments/scaling) | Réplicas e infraestrutura |
| [Railway — serverless](https://docs.railway.com/deployments/serverless) | Verificação da política de hibernação |
| [Alpaca — paper trading](https://docs.alpaca.markets/us/docs/paper-trading) | Ambiente candidato e limites da simulação |
| [Alpaca — ordens](https://docs.alpaca.markets/us/docs/working-with-orders) | Client order ID e consulta de ordens |

Alpaca paper usa endpoint/credenciais próprios; paper não substitui validação em
mercado real. Contas paper-only têm limitações de feed; validar o plano escolhido.
A documentação descreve consulta por `client_order_id`, mas isso não oferece por
si só garantia de execução exatamente uma vez. Não presumir suporte a B3.

Código local consultado: `bootstrap/app.go`, `routes/web.go`, `go.mod`,
`package.json`, `Dockerfile`, CI e scaffold upstream Vukem em `/tmp/vukem`.
A cópia upstream serve para investigação; ao implementar, verificar as mesmas
APIs no módulo 0.37.0 fixado em `go.mod`, pois `main` pode evoluir.

O texto de contexto contém valores de custos, prazos e alegações regulatórias.
Eles não foram adotados como cotações ou pareceres verificados. Para SaaS/live,
registrar pesquisa e revisão específicas de jurisdição e serviço antes da oferta.

Verificação adicional conferiu social login e middleware também no módulo Vukem
0.37.0 instalado em `GOPATH/pkg/mod`, não apenas no upstream main. Railway
pre-deploy usa container separado, rede/env disponíveis, sem volumes montados;
start command Docker substitui ENTRYPOINT. Referências adicionais:
[Start command](https://docs.railway.com/deployments/start-command) e
[Serviços](https://docs.railway.com/services).
