# ADRs — decisões de arquitetura

Estados: **aceito** quando deriva de requisito explícito; **proposto** quando
precisa de aceite do usuário. Aceitar ADR não significa implementá-lo.
ADRs substituídos permanecem no histórico, com link para o sucessor.

| ADR | Tema | Estado |
| --- | --- | --- |
| [0001](0001-stack-e-dashboard.md) | Stack, dashboard e shadcn/ui | Aceito por requisito; parcialmente implementado |
| [0002](0002-monolito-modular-worker.md) | Monólito modular com worker separado | Proposto |
| [0003](0003-google-only.md) | Login exclusivo Google | Aceito por requisito; implementação pendente |
| [0004](0004-postgres-outbox.md) | PostgreSQL, outbox e precisão financeira | Proposto; PostgreSQL já confirmado |
| [0005](0005-railway.md) | Deploy Railway | Aceito por requisito; topologia proposta |
| [0006](0006-dados-e-eventos.md) | Dados, extensões e mensageria | Proposto |
| [0007](0007-paper-first.md) | Paper-first e execução determinística | Paper-first aceito; detalhes propostos |
| [0008](0008-brasil-e-eua.md) | Brasil e EUA no MVP | Aceito; provedores pendentes |
| [0009](0009-uso-pessoal.md) | Terminal pessoal e conta única | Aceito |
| [0010](0010-adapters-paper.md) | Alpaca Paper EUA e simulador B3 | Implementados; validação externa pendente |
| [0011](0011-implementacao-autonoma.md) | Defaults e implementação autônoma | Aceito por instrução posterior |
| [0012](0012-snapshots-transacionais.md) | Persistência efetivamente implementada | Implementado; detalha/atualiza ADR-0004 |
| [0013](0013-workers-polling-e-deploy.md) | Polling, liderança e standby Railway | Implementado |
| [0014](0014-railway-api-e-docker-portavel.md) | API Railway atual e Dockerfile portátil | Implementado durante publicação |

Cada ADR deve registrar contexto, decisão, alternativas, consequências e gatilho
de revisão. Brasil e EUA e uso pessoal estão confirmados; Alpaca Paper, simulador interno B3 e admissão de uma conta também. IEX e CSV B3 foram adotados por delegação; credenciais/datasets continuam externos.
