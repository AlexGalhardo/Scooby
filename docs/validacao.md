# Validação técnica, quantitativa e operacional

> Atualização de implementação: consulte o [ROADMAP](../ROADMAP.md),
> [ADR-0011](adrs/0011-implementacao-autonoma.md),
> [ADR-0012](adrs/0012-snapshots-transacionais.md) e
> [ADR-0013](adrs/0013-workers-polling-e-deploy.md) para o estado efetivo.
> O texto abaixo preserva requisitos/desenhos originais, não uma declaração de
> que todos os detalhes de infraestrutura ou validação externa já ocorreram.

Estado: plano de testes; não evidência de implementação.

## Autenticação e dashboard

Visitante em `/` recebe redirect a `/login` sem dados privados. `/login` contém
apenas o provedor Google, sem senha, cadastro local ou reset. Usuário autenticado
volta a `/`. Recusar callback sem state, expirado, reapresentado ou com token
inválido. Identidade não autorizada não cria sessão válida. Testar fixação de
sessão, logout, revogação, CSRF e redirect externo. Routes antigas e endpoints
JSON/stream obedecem às mesmas permissões; dois usuários não acessam conta alheia.
Sidebar funciona por teclado e em mobile; contrastes, loading, vazio, indisponível
e dados atrasados são distintos. Emergência pede confirmação de ação destrutiva,
mas o bloqueio rápido de novas entradas deve ser acessível e inequívoco.

## Domínio e motor de risco

Testes tabelados de precisão, tick/lote, compra/saída, stop inválido, custo,
slippage, exposição, perda diária e preço vencido. Duas intenções concorrentes não
consomem a mesma margem de risco. Risco de ordens abertas/unknown permanece
reservado. O dia de negociação vem do calendário do mercado; horário local do
servidor não redefine o limite. Reiniciar não zera perda diária ou kill switch.

## Matriz de falhas da execução

| Falha injetada | Resultado esperado |
| --- | --- |
| Crash depois de gravar intenção, antes do envio | Recuperar intenção sem duplicar |
| Timeout após broker aceitar ordem | Estado unknown; consultar por identificador antes de ação |
| Fill duplicado ou fora de ordem | Uma aplicação no ledger, projeção coerente |
| Execução parcial seguida de cancelamento | Atualizar quantidade e risco remanescente corretamente |
| Feed cai, atrasa ou contém outlier | Bloquear novas entradas, sinalizar degradação |
| Banco indisponível | Interromper novas decisões; reconciliar antes de retomar |
| Deploy com dois workers sobrepostos | Geração/lock impedem novos envios do antigo; reconciliar em voo |
| Cancelamento sem confirmação | Não presumir encerrado nem liberar reserva |
| Kill switch com ordem em voo | Bloqueio durável e resultado externo reconciliado |
| Sessão/feriado/DST | Sem operação fora da política do instrumento |
| Posição externa diverge | Quarentena da conta; retomada controlada |

## Evidência quantitativa

Reservar conjuntos por tempo: desenvolvimento, validação e teste final intocado.
Walk-forward refaz seleção apenas com passado disponível. Registrar sobrevivência
dos ativos, splits, dividendos, timezone, ajustes, gaps e revisões. Separar preço
ajustado para pesquisa do preço negociável para fills; registrar conversões.
Proibir sinal pelo fechamento do candle com fill no mesmo fechamento já conhecido;
no padrão, executar no próximo evento elegível com custos.

Relatar retorno líquido, drawdown, duração de drawdown, turnover, número de
operações, exposição, ganho/perda médios, win rate e sensibilidade a custos.
Comparar benchmark e taxas na mesma moeda/período. Sharpe só com frequência,
anualização e hipóteses explícitas. Poucas operações e múltiplas tentativas de
parâmetros reduzem confiança; reportar isso, não apenas a melhor curva.

## Gates de operação

Backtest reproduzível → simulação com falhas → paper observado → revisão.
Critérios numéricos (perda máxima, divergência, tolerância de atraso, período e
amostra mínima) serão pactuados no grilling. Paper não reproduz integralmente
liquidez, fila, taxas ou slippage real. Aprovação de desempenho técnico não é
aprovação financeira. Live depende de decisão futura e novo aceite de critérios.

## Ferramentas e CI propostos

Manter `go test -race`, `go vet`, `govulncheck`, TypeScript, build, secret scanning
e imagem. Acrescentar contratos dos adapters com fixtures, PostgreSQL real,
fluxo OAuth fake em teste, navegação E2E, restore e replay de cenários de falha.
Sem credenciais reais de corretora no CI; testes externos somente em paper e
ambiente dedicado. A mudança documental atual exige links e consistência, não
reexecutar toda a suíte do starter sem mudança de código.
