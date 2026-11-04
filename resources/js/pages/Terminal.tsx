import { Head } from "@inertiajs/react";
import {
  Activity,
  ArrowDownToLine,
  ArrowUpRight,
  CircleHelp,
  Database,
  FlaskConical,
  Loader2,
  Plus,
  RefreshCw,
  ShieldAlert,
  Upload,
  Wifi,
  WifiOff,
} from "lucide-react";
import {
  cloneElement,
  type FormEvent,
  isValidElement,
  type ReactElement,
  type ReactNode,
  useCallback,
  useEffect,
  useState,
} from "react";
import MarketChart from "@/components/MarketChart";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import { api } from "@/lib/api";
import {
  type Account,
  alive,
  type Candle,
  type Dataset,
  date,
  type Job,
  money,
  type Policy,
  type Snapshot,
  type Strategy,
} from "@/types";

const titles: Record<string, [string, string]> = {
  overview: ["Visão geral", "Seu mercado. Suas regras. Todas as decisões em um só lugar."],
  market: ["Dados de mercado", "Histórico com origem identificada. Sem dados inventados."],
  strategies: ["Estratégias", "Transforme hipóteses em regras determinísticas e versionadas."],
  backtests: ["Laboratório", "Teste, compare e avalie depois dos custos."],
  orders: ["Ordens", "Do sinal à confirmação. Cada transição importa."],
  positions: ["Posições", "Exposição segregada por conta, moeda e modalidade."],
  risk: ["Controle de risco", "Defina limites antes de permitir novas exposições."],
  audit: ["Auditoria", "Uma trilha persistente de decisões e ações."],
  settings: ["Configurações", "Status das integrações e do ambiente de execução."],
};
const kindName: Record<string, string> = {
  ma_cross: "Cruzamento de médias",
  mean_reversion: "Reversão à média",
  breakout: "Rompimento",
};
function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <Label className="grid gap-2 text-xs font-normal text-muted-foreground">
      {label}
      {isValidElement(children)
        ? cloneElement(children as ReactElement<{ "aria-label"?: string }>, { "aria-label": label })
        : children}
    </Label>
  );
}
function Select({
  value,
  onChange,
  children,
  ...props
}: {
  value: string;
  onChange: (v: string) => void;
  children: ReactNode;
  "aria-label"?: string;
}) {
  return (
    <select
      {...props}
      className="h-9 w-full rounded-lg border bg-background px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-emerald-500"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    >
      {children}
    </select>
  );
}
function Empty({ title, detail }: { title: string; detail: string }) {
  return (
    <div className="flex min-h-48 flex-col items-center justify-center gap-3 p-8 text-center">
      <div className="rounded-xl border bg-muted/30 p-3">
        <Database className="size-5 text-muted-foreground" />
      </div>
      <p className="text-sm font-medium">{title}</p>
      <p className="max-w-md text-xs leading-relaxed text-muted-foreground">{detail}</p>
    </div>
  );
}
function Status({ value }: { value: string }) {
  const good = ["enabled", "completed", "filled", "ready"].includes(value);
  return (
    <Badge
      variant="outline"
      className={good ? "border-emerald-500/20 text-emerald-500" : "text-muted-foreground"}
    >
      {value}
    </Badge>
  );
}
interface GridRow {
  id: string | number;
  cells: ReactNode[];
}

const row = (id: string | number, ...cells: ReactNode[]): GridRow => ({ id, cells });

function GridTable({ headers, rows }: { headers: string[]; rows: GridRow[] }) {
  return rows.length ? (
    <div className="overflow-auto">
      <Table>
        <TableHeader>
          <TableRow>
            {headers.map((h) => (
              <TableHead key={h} className="text-[11px] uppercase tracking-wide">
                {h}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((r) => (
            <TableRow key={r.id}>
              {r.cells.map((cell, j) => (
                <TableCell key={headers[j]} className="text-xs">
                  {cell}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  ) : (
    <Empty
      title="Nenhum registro ainda"
      detail="Os dados aparecerão aqui depois que você configurar uma estratégia e executar uma simulação."
    />
  );
}
function Stat({
  label,
  value,
  detail,
  positive,
}: {
  label: string;
  value: string;
  detail: string;
  positive?: boolean;
}) {
  return (
    <Card>
      <CardContent className="pt-5">
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          {label}
          <ArrowUpRight className="size-3" />
        </div>
        <p
          className={`mt-3 font-mono text-2xl tracking-tight ${positive ? "text-emerald-500" : ""}`}
        >
          {value}
        </p>
        <p className="mt-2 text-[10px] text-muted-foreground">{detail}</p>
      </CardContent>
    </Card>
  );
}
export default function Terminal({ section }: { section: string }) {
  const [snapshot, setSnapshot] = useState<Snapshot>();
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [busy, setBusy] = useState(false);
  const [accountId, setAccount] = useState("b3");
  const [datasetId, setDataset] = useState("");
  const [candles, setCandles] = useState<Candle[]>([]);
  const [report, setReport] = useState<Job>();
  const [reportOpen, setReportOpen] = useState(false);
  const refresh = useCallback(async () => {
    try {
      setSnapshot(await api<Snapshot>("/api/snapshot"));
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Falha de conexão");
    }
  }, []);
  useEffect(() => {
    void refresh();
    const timer = setInterval(() => {
      if (!document.hidden) void refresh();
    }, 10000);
    return () => clearInterval(timer);
  }, [refresh]);
  const act = async (path: string, body: unknown) => {
    setBusy(true);
    setNotice("");
    try {
      await api(path, body);
      await refresh();
      setNotice("Solicitação registrada.");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Falha ao executar ação");
    } finally {
      setBusy(false);
    }
  };
  useEffect(() => {
    if (!datasetId) {
      setCandles([]);
      return;
    }
    const controller = new AbortController();
    api<Dataset>(`/api/datasets/${datasetId}`, undefined, controller.signal)
      .then((ds) => setCandles(ds.candles ?? []))
      .catch((e) => {
        if (!controller.signal.aborted) setError(String(e));
      });
    return () => controller.abort();
  }, [datasetId]);
  const account = snapshot?.accounts.find((a) => a.id === accountId);
  const [title, description] = titles[section] ?? titles.overview;
  if (!snapshot)
    return (
      <>
        <Head title={title} />
        <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4">
          {error ? (
            <>
              <WifiOff />
              <p role="alert">{error}</p>
              <Button onClick={() => void refresh()}>Tentar novamente</Button>
            </>
          ) : (
            <>
              <Loader2 className="animate-spin text-emerald-400" />
              <p className="text-sm text-muted-foreground">Carregando seu workspace…</p>
            </>
          )}
        </div>
      </>
    );
  const currency = account?.currency ?? "BRL";
  const known = account?.id === "b3" || (!!account && !account.reconciled_at.startsWith("0001"));
  const controls = (a: Account) => (
    <div className="flex flex-wrap gap-2">
      <Confirm
        title={a.halted ? "Retomar simulação" : "Bloquear novas entradas"}
        description={
          a.halted
            ? "A engine continuará operando sem o navegador. Revise dados e limites antes de retomar."
            : "Bloqueia novas entradas; ordens em voo e posições ainda precisam ser acompanhadas."
        }
        label={a.halted ? "Retomar PAPER" : "Bloquear"}
        disabled={busy}
        action={() =>
          act(`/api/accounts/${a.id}/control`, {
            action: a.halted ? "resume" : "halt",
            confirm: "PAPER",
          })
        }
      />
      {a.id === "b3" && (
        <Confirm
          title="Novo replay B3"
          description="Arquiva o snapshot anterior, preservando fills e auditoria, e reinicia o capital fictício em R$ 100.000. Exige conta bloqueada e sem posições/ordens abertas."
          label="Novo replay"
          disabled={busy}
          action={() => act(`/api/accounts/${a.id}/control`, { action: "reset", confirm: "RESET" })}
        />
      )}
      <Confirm
        title="Cancelar ordens abertas"
        description="A conta será bloqueada. Cancelamentos externos permanecem pendentes até confirmação."
        label="Cancelar ordens"
        disabled={busy}
        action={() => act(`/api/accounts/${a.id}/control`, { action: "cancel" })}
      />
      {a.id === "us" && (
        <Confirm
          title="Zerar posições paper"
          description="Cria ordens de saída com base na posição reconciliada. Não garante execução nem encerramento imediato."
          label="Zerar posições"
          disabled={busy}
          action={() =>
            act(`/api/accounts/${a.id}/control`, { action: "flatten", confirm: "ZERAR" })
          }
        />
      )}
    </div>
  );
  return (
    <>
      <Head title={`${title} · Scooby`} />
      <div className="mb-7 flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="mb-2 text-[10px] font-medium uppercase tracking-[.2em] text-emerald-500">
            Workspace / {section === "backtests" ? "Pesquisa quantitativa" : "Trading terminal"}
          </p>
          <h1 className="text-3xl font-semibold tracking-tight">{title}</h1>
          <p className="mt-2 text-sm text-muted-foreground">{description}</p>
        </div>
        <div className="flex items-center gap-2">
          <Select value={accountId} onChange={setAccount}>
            <option value="b3">Brasil · BRL · Replay</option>
            <option value="us">EUA · USD · Alpaca Paper</option>
          </Select>
          <Button
            size="icon"
            variant="outline"
            onClick={() => void refresh()}
            aria-label="Atualizar"
          >
            <RefreshCw className="size-4" />
          </Button>
        </div>
      </div>
      {error && (
        <div
          role="alert"
          className="mb-4 flex justify-between rounded-xl border border-red-500/30 bg-red-500/5 p-4 text-sm text-red-500"
        >
          {error}
          <button type="button" onClick={() => setError("")} aria-label="Fechar alerta">
            ×
          </button>
        </div>
      )}
      {notice && (
        <div
          role="status"
          className="mb-4 rounded-xl border border-emerald-500/20 p-3 text-xs text-emerald-500"
        >
          {notice}
        </div>
      )}
      {account && (
        <div className="mb-6 flex flex-wrap items-center gap-3 rounded-xl border bg-muted/15 px-4 py-3 text-xs">
          <span
            className={`size-1.5 rounded-full ${alive(account.heartbeat) ? "bg-emerald-400" : "bg-amber-400"}`}
          />
          <span>{alive(account.heartbeat) ? "Worker conectado" : "Aguardando worker"}</span>
          <span className="text-muted-foreground">/</span>
          <Badge variant="secondary">
            {account.mode === "simulation" ? "SIMULAÇÃO B3 · REPLAY" : "ALPACA PAPER · IEX"}
          </Badge>
          <span className="ml-auto text-muted-foreground">
            {account.halted ? account.reason : "Novas entradas habilitadas"}
          </span>
        </div>
      )}
      {section === "overview" && account && (
        <div className="space-y-6">
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <Stat
              label="Patrimônio simulado"
              value={
                account.id === "us" && account.reconciled_at.startsWith("0001")
                  ? "—"
                  : money(account.equity, currency)
              }
              detail={`${currency} · conta segregada`}
            />
            <Stat
              label="P&L realizado"
              value={known ? money(account.realized, currency) : "—"}
              detail="Com base nos fills registrados"
              positive={Number(account.realized) >= 0}
            />
            <Stat
              label="Exposição reservada + posições"
              value={known ? money(account.exposure, currency) : "—"}
              detail="Inclui ordens pendentes"
            />
            <Stat
              label="Drawdown atual"
              value={known ? `${Number(account.drawdown).toFixed(2)}%` : "—"}
              detail="Em relação ao pico observado"
            />
          </div>
          <div className="grid gap-5 xl:grid-cols-[1fr_320px]">
            <Card>
              <CardHeader className="flex-row items-start justify-between">
                <div>
                  <CardTitle>Explorar o histórico</CardTitle>
                  <CardDescription>
                    Até 1.000 candles recentes do dataset selecionado
                  </CardDescription>
                </div>
                <div className="max-w-48">
                  <DatasetSelect snapshot={snapshot} value={datasetId} onChange={setDataset} />
                </div>
              </CardHeader>
              <CardContent>
                {candles.length ? (
                  <MarketChart candles={candles} />
                ) : (
                  <Empty
                    title="Seu gráfico começa com dados"
                    detail="Importe um histórico autorizado ou consulte a Alpaca Paper em Dados de mercado."
                  />
                )}
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Estado operacional</CardTitle>
                <CardDescription>Controle independente do navegador</CardDescription>
              </CardHeader>
              <CardContent className="space-y-5">
                <div className="flex items-center gap-3">
                  {account.halted ? (
                    <ShieldAlert className="text-amber-500" />
                  ) : (
                    <Activity className="text-emerald-500" />
                  )}
                  <div>
                    <p className="text-sm">
                      {account.halted ? "Entradas bloqueadas" : "Simulação ativa"}
                    </p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {account.reason || "Monitoramento pelo worker Go"}
                    </p>
                  </div>
                </div>
                <div className="space-y-2 border-y py-4 text-xs">
                  <p className="flex justify-between">
                    <span className="text-muted-foreground">Estratégias habilitadas</span>
                    {
                      snapshot.strategies.filter(
                        (s) => s.status === "enabled" && s.payload.account_id === account.id,
                      ).length
                    }
                  </p>
                  <p className="flex justify-between">
                    <span className="text-muted-foreground">Caixa</span>
                    {known ? money(account.cash, currency) : "—"}
                  </p>
                  <p className="flex justify-between">
                    <span className="text-muted-foreground">P&L não realizado</span>
                    {known ? money(account.unrealized, currency) : "—"}
                  </p>
                  <p className="text-muted-foreground">Heartbeat: {date(account.heartbeat)}</p>
                </div>
                {controls(account)}
                <p className="text-[11px] leading-relaxed text-muted-foreground">
                  IEX não representa o volume consolidado americano. Replay B3 usa dados históricos
                  e fills internos.
                </p>
              </CardContent>
            </Card>
          </div>
          <Card>
            <CardHeader>
              <CardTitle>Últimas ordens</CardTitle>
            </CardHeader>
            <CardContent>
              <Orders account={account} act={act} />
            </CardContent>
          </Card>
        </div>
      )}
      {section === "market" && (
        <div className="space-y-5">
          <div className="grid gap-5 xl:grid-cols-2">
            <ImportForm snapshot={snapshot} refresh={refresh} setError={setError} />
            <Card>
              <CardHeader>
                <CardTitle>Histórico americano · IEX</CardTitle>
                <CardDescription>
                  Usa as credenciais Alpaca no servidor. Respeita a janela recente restrita do plano
                  básico.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <FetchForm configured={snapshot.settings.alpaca_configured} act={act} busy={busy} />
                <p className="mt-5 text-xs leading-relaxed text-muted-foreground">
                  Nenhuma assinatura é contratada automaticamente. A fonte B3 em tempo real ainda
                  não está integrada.
                </p>
              </CardContent>
            </Card>
          </div>
          <Card>
            <CardHeader>
              <CardTitle>Datasets disponíveis</CardTitle>
              <CardDescription>Preservados com hash, origem, licença e período</CardDescription>
            </CardHeader>
            <CardContent>
              <GridTable
                headers={["Ativo", "Fonte", "Candles", "Período", "Qualidade", "Visualizar"]}
                rows={snapshot.datasets.map((ds) =>
                  row(
                    ds.id,
                    <span className="font-mono">{ds.symbol}</span>,
                    ds.source,
                    ds.bars,
                    <span>
                      {date(ds.first)}
                      <br />
                      {date(ds.last)}
                    </span>,
                    ds.problems?.length ? (
                      <Badge variant="outline">{ds.problems.length} gaps</Badge>
                    ) : (
                      <Status value="ready" />
                    ),
                    <Button size="sm" variant="outline" onClick={() => setDataset(ds.id)}>
                      Gráfico
                    </Button>,
                  ),
                )}
              />
              {candles.length > 0 && (
                <div className="mt-6">
                  <p className="mb-3 text-xs text-muted-foreground">
                    {snapshot.datasets.find((d) => d.id === datasetId)?.symbol} · últimos{" "}
                    {candles.length} candles
                  </p>
                  <MarketChart candles={candles} />
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}
      {section === "strategies" && (
        <div className="space-y-5">
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">
              Cada configuração é imutável; crie nova versão para alterar regras.
            </p>
            <Dialog>
              <DialogTrigger asChild>
                <Button>
                  <Plus />
                  Nova estratégia
                </Button>
              </DialogTrigger>
              <DialogContent className="max-h-[90vh] overflow-auto">
                <DialogHeader>
                  <DialogTitle>Nova configuração</DialogTitle>
                  <DialogDescription>
                    Começa desabilitada. Ativar a estratégia não desbloqueia a conta.
                  </DialogDescription>
                </DialogHeader>
                <StrategyForm snapshot={snapshot} refresh={refresh} />
              </DialogContent>
            </Dialog>
          </div>
          {snapshot.strategies.length ? (
            <div className="grid gap-5 lg:grid-cols-2 xl:grid-cols-3">
              {snapshot.strategies.map(({ payload: s, status }) => (
                <Card key={s.id}>
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <Badge variant="outline">
                        {s.symbol} · {s.account_id.toUpperCase()}
                      </Badge>
                      <Status value={status} />
                    </div>
                    <CardTitle className="pt-3">{s.name}</CardTitle>
                    <CardDescription>
                      {kindName[s.kind]} · versão {s.version}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="mb-5 grid grid-cols-2 gap-3 rounded-xl bg-muted/30 p-4 text-xs">
                      <span className="text-muted-foreground">Janelas</span>
                      <span>
                        {s.fast} / {s.slow}
                      </span>
                      <span className="text-muted-foreground">Stop / alvo</span>
                      <span>
                        {s.stop_bps / 100}% / {s.target_bps / 100}%
                      </span>
                      <span className="text-muted-foreground">Timeframe</span>
                      <span>5 minutos · sessão regular</span>
                    </div>
                    <Button
                      className="w-full"
                      variant={status === "enabled" ? "outline" : "default"}
                      disabled={busy}
                      onClick={() =>
                        void act(`/api/strategies/${s.id}/toggle`, {
                          enabled: status !== "enabled",
                        })
                      }
                    >
                      {status === "enabled" ? "Desabilitar" : "Habilitar estratégia"}
                    </Button>
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : (
            <Card>
              <Empty
                title="Qual hipótese você quer testar?"
                detail="Comece com médias móveis, reversão à média ou rompimento. Parâmetros e limites são validados no servidor."
              />
            </Card>
          )}
        </div>
      )}
      {section === "backtests" && (
        <div className="space-y-5">
          <Card>
            <CardHeader>
              <CardTitle>Executar experimento</CardTitle>
              <CardDescription>
                Job durável processado pelo worker. Treino e teste têm divisão temporal de 70% /
                30%.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <BacktestForm snapshot={snapshot} act={act} busy={busy} />
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Histórico de experimentos</CardTitle>
              <CardDescription>
                Compare resultados líquidos e limites da amostra; não ajuste parâmetros no teste
                final.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <GridTable
                headers={[
                  "Estratégia",
                  "Partição",
                  "Status",
                  "Retorno",
                  "Drawdown",
                  "Operações",
                  "Relatório",
                ]}
                rows={snapshot.backtests.map(({ id, payload: j, status }) =>
                  row(
                    id,
                    j.strategy.name,
                    j.partition,
                    <Status value={status} />,
                    j.result ? `${Number(j.result.metrics.return_pct).toFixed(2)}%` : "—",
                    j.result ? `${Number(j.result.metrics.max_drawdown_pct).toFixed(2)}%` : "—",
                    j.result?.metrics.trades ?? "—",
                    <div className="flex gap-2">
                      {j.result && (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={async () => {
                            try {
                              setReport(await api<Job>(`/api/backtests/${j.id}`));
                              setReportOpen(true);
                            } catch (e) {
                              setError(String(e));
                            }
                          }}
                        >
                          Abrir
                        </Button>
                      )}
                      {["queued", "running"].includes(status) && (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => void act(`/api/backtests/${j.id}/cancel`, {})}
                        >
                          Cancelar
                        </Button>
                      )}
                      {j.error && <span className="text-red-400">{j.error}</span>}
                    </div>,
                  ),
                )}
              />
            </CardContent>
          </Card>
          <Dialog open={reportOpen} onOpenChange={setReportOpen}>
            <DialogContent className="max-h-[90vh] overflow-auto sm:max-w-4xl">
              <DialogHeader>
                <DialogTitle>{report?.strategy.name} · relatório</DialogTitle>
                <DialogDescription>
                  Equity líquida e benchmark buy-and-hold. Partição: {report?.partition}
                </DialogDescription>
              </DialogHeader>
              {report?.result && (
                <>
                  <MarketChart curve={report.result.curve ?? []} />
                  <div className="grid gap-3 sm:grid-cols-3">
                    <Stat
                      label="Custos simulados"
                      value={money(
                        report.result.metrics.costs,
                        report.strategy.account_id === "b3" ? "BRL" : "USD",
                      )}
                      detail="Taxas configuradas em bps"
                    />
                    <Stat
                      label="Win rate"
                      value={`${report.result.metrics.win_rate.toFixed(1)}%`}
                      detail={`${report.result.metrics.trades} operações encerradas`}
                    />
                    <Stat
                      label="Exposição"
                      value={String(report.result.metrics.exposure_bars)}
                      detail="Candles com posição"
                    />
                  </div>
                  <div className="space-y-2 text-xs text-muted-foreground">
                    {report.result.warnings.map((w) => (
                      <p key={w}>• {w}</p>
                    ))}
                    <p className="break-all">Dataset SHA256: {report.result.dataset_hash}</p>
                  </div>
                  <Button
                    variant="outline"
                    onClick={() => {
                      const blob = new Blob([JSON.stringify(report, null, 2)], {
                        type: "application/json",
                      });
                      const href = URL.createObjectURL(blob);
                      const a = document.createElement("a");
                      a.href = href;
                      a.download = `scooby-backtest-${report.id}.json`;
                      a.click();
                      URL.revokeObjectURL(href);
                    }}
                  >
                    <ArrowDownToLine />
                    Exportar artefato JSON
                  </Button>
                  <GridTable
                    headers={["Horário", "Ativo", "Lado", "Quantidade", "Preço", "Taxas"]}
                    rows={(report.result.fills ?? [])
                      .slice(-100)
                      .map((f) =>
                        row(f.id, date(f.time), f.symbol, f.side, f.qty, f.price, f.fees),
                      )}
                  />
                </>
              )}
            </DialogContent>
          </Dialog>
        </div>
      )}
      {section === "orders" && account && (
        <Card>
          <CardHeader>
            <CardTitle>Ordens · {account.id.toUpperCase()}</CardTitle>
            <CardDescription>
              Últimos 100 registros. Estado desconhecido mantém a reserva de risco.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Orders account={account} act={act} />
          </CardContent>
        </Card>
      )}
      {section === "positions" && account && (
        <div className="space-y-5">
          <Card>
            <CardHeader>
              <CardTitle>Posições abertas</CardTitle>
              <CardDescription>
                Marcação atual e proteção definidas pela intenção de entrada
              </CardDescription>
            </CardHeader>
            <CardContent>
              <GridTable
                headers={[
                  "Ativo",
                  "Quantidade",
                  "Preço médio",
                  "Marcação",
                  "Stop",
                  "Alvo",
                  "P&L não realizado",
                ]}
                rows={Object.values(account.positions).map((p) =>
                  row(
                    p.symbol,
                    p.symbol,
                    p.qty,
                    money(p.average, currency),
                    money(p.mark, currency),
                    money(p.stop, currency),
                    money(p.target, currency),
                    money((Number(p.mark) - Number(p.average)) * Number(p.qty), currency),
                  ),
                )}
              />
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Execuções registradas</CardTitle>
              <CardDescription>
                Valores de exibição; a contabilidade usa decimais no servidor
              </CardDescription>
            </CardHeader>
            <CardContent>
              <GridTable
                headers={["Horário", "Ativo", "Lado", "Quantidade", "Preço", "Taxas"]}
                rows={[...account.fills]
                  .reverse()
                  .map((f) =>
                    row(
                      f.id,
                      date(f.time),
                      f.symbol,
                      f.side,
                      f.qty,
                      money(f.price, currency),
                      money(f.fees, currency),
                    ),
                  )}
              />
            </CardContent>
          </Card>
        </div>
      )}
      {section === "risk" && account && (
        <div className="grid gap-5 xl:grid-cols-[1fr_360px]">
          <Card>
            <CardHeader>
              <CardTitle>Política de risco · {account.id.toUpperCase()}</CardTitle>
              <CardDescription>
                Bloqueie a conta para editar. Limites persistem entre reinícios.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <RiskForm key={account.id} account={account} act={act} busy={busy} />
            </CardContent>
          </Card>
          <Card className="border-amber-500/20">
            <CardHeader>
              <ShieldAlert className="mb-2 text-amber-500" />
              <CardTitle>Controles operacionais</CardTitle>
              <CardDescription>
                Bloquear novas entradas não significa que posições foram encerradas.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              {controls(account)}
              <p className="text-xs leading-relaxed text-muted-foreground">
                Ordens em voo precisam de confirmação. Stops e alvos são observados pela engine, não
                garantem preço ou proteção durante indisponibilidade. Para B3, o replay aplica
                regras no relógio histórico.
              </p>
              <p className="text-xs text-muted-foreground">
                Reconciliação: {date(account.reconciled_at)}
              </p>
            </CardContent>
          </Card>
        </div>
      )}
      {section === "audit" && (
        <Card>
          <CardHeader>
            <CardTitle>Eventos auditados</CardTitle>
            <CardDescription>Últimos 100 eventos persistidos</CardDescription>
          </CardHeader>
          <CardContent>
            <GridTable
              headers={["Horário", "Ator", "Ação", "Conta / registro", "Detalhes"]}
              rows={snapshot.audit.map((a) =>
                row(
                  a.id,
                  date(a.time),
                  a.actor,
                  <code>{a.action}</code>,
                  <span className="font-mono text-[10px]">{a.account}</span>,
                  a.action === "simulation.reset" ? (
                    <a
                      className="text-emerald-500 underline"
                      href={`/api/artifacts/${encodeURIComponent(a.details)}`}
                      target="_blank"
                      rel="noreferrer"
                    >
                      Snapshot anterior (JSON)
                    </a>
                  ) : (
                    a.details
                  ),
                ),
              )}
            />
          </CardContent>
        </Card>
      )}
      {section === "settings" && (
        <div className="grid gap-5 lg:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Integrações</CardTitle>
              <CardDescription>
                Somente presença de configuração; segredos nunca são retornados ao navegador
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              {[
                ["Login Google", snapshot.settings.google_configured],
                ["Proprietário autorizado", snapshot.settings.owner_configured],
                ["Alpaca Paper", snapshot.settings.alpaca_configured],
              ].map(([label, ready]) => (
                <div key={String(label)} className="flex justify-between border-b pb-3 text-sm">
                  <span>{label}</span>
                  <Status value={ready ? "ready" : "não configurado"} />
                </div>
              ))}
              <p className="text-xs text-muted-foreground">
                Configure credenciais e callback pelo ambiente Railway. B3 usa importação de
                histórico autorizado.
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Contrato do ambiente</CardTitle>
              <CardDescription>Escopo pessoal e exclusivamente simulado</CardDescription>
            </CardHeader>
            <CardContent>
              <GridTable
                headers={["Configuração", "Valor"]}
                rows={[
                  row("mode", "Modo", "Paper only"),
                  row("feed", "Feed americano", "IEX"),
                  row("timeframe", "Timeframe", "5 minutos"),
                  row("engine", "Engine", snapshot.settings.engine_version),
                  row("persistence", "Persistência", "PostgreSQL"),
                  row("server-time", "Horário servidor", date(snapshot.server_time)),
                ]}
              />
              <div className="mt-5 flex gap-3 rounded-lg bg-muted/40 p-4 text-xs leading-relaxed text-muted-foreground">
                <CircleHelp className="size-4 shrink-0" />A recuperação da conta requer procedimento
                administrativo auditado. Alterar o email de configuração não substitui
                automaticamente uma identidade já vinculada.
              </div>
            </CardContent>
          </Card>
        </div>
      )}
      <p className="mt-6 flex items-center gap-2 text-[10px] text-muted-foreground">
        <Wifi className="size-3" />
        Snapshot: {date(snapshot.server_time)} · atualização a cada 10 segundos · {currency}
      </p>
    </>
  );
}
function DatasetSelect({
  snapshot,
  value,
  onChange,
  ...props
}: {
  snapshot: Snapshot;
  value: string;
  onChange: (v: string) => void;
  "aria-label"?: string;
}) {
  return (
    <Select {...props} value={value} onChange={onChange}>
      <option value="">Selecione histórico</option>
      {snapshot.datasets.map((d) => (
        <option key={d.id} value={d.id}>
          {d.symbol} · {d.bars} candles · {d.source}
        </option>
      ))}
    </Select>
  );
}
function Confirm({
  title,
  description,
  label,
  disabled,
  action,
}: {
  title: string;
  description: string;
  label: string;
  disabled: boolean;
  action: () => Promise<void>;
}) {
  const [open, setOpen] = useState(false);
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" disabled={disabled}>
          {label}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setOpen(false)}>
            Voltar
          </Button>
          <Button
            disabled={disabled}
            onClick={async () => {
              await action();
              setOpen(false);
            }}
          >
            Confirmar
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
function Orders({
  account,
  act,
}: {
  account: Account;
  act: (p: string, b: unknown) => Promise<void>;
}) {
  return (
    <GridTable
      headers={["Horário", "Ativo", "Lado", "Qtd / executada", "Limite", "Estado", "Ação"]}
      rows={[...account.orders].reverse().map((o) =>
        row(
          o.id,
          date(o.created_at),
          <span className="font-mono">{o.symbol}</span>,
          <Badge variant="outline">{o.side === "buy" ? "Compra" : "Venda"}</Badge>,
          `${o.qty} / ${o.filled}`,
          money(o.limit, account.currency),
          <span>
            <Status value={o.status} />
            {o.reason && (
              <span className="mt-1 block max-w-48 text-[10px] text-muted-foreground">
                {o.reason}
              </span>
            )}
          </span>,
          ["approved", "dispatching", "unknown", "acknowledged", "partially_filled"].includes(
            o.status,
          ) ? (
            <Button
              size="sm"
              variant="outline"
              onClick={() => void act(`/api/accounts/${account.id}/orders/${o.id}/cancel`, {})}
            >
              Cancelar
            </Button>
          ) : (
            "—"
          ),
        ),
      )}
    />
  );
}
function ImportForm({
  snapshot,
  refresh,
  setError,
}: {
  snapshot: Snapshot;
  refresh: () => Promise<void>;
  setError: (v: string) => void;
}) {
  const [symbol, setSymbol] = useState("PETR4");
  const [source, setSource] = useState("");
  const [license, setLicense] = useState("");
  const [csv, setCsv] = useState("");
  const [busy, setBusy] = useState(false);
  const [success, setSuccess] = useState("");
  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    try {
      const result = await api<{ bars: number }>("/api/datasets/import", {
        symbol,
        source,
        license,
        csv,
      });
      setSuccess(`${result.bars} candles importados com hash e procedência.`);
      setCsv("");
      await refresh();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <Card>
      <CardHeader>
        <CardTitle>Importar histórico CSV</CardTitle>
        <CardDescription>
          B3 2026 · timestamp com offset; time,open,high,low,close,volume
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={submit} className="grid gap-4">
          <div className="grid gap-3 sm:grid-cols-2">
            <Field label="Ativo">
              <Select value={symbol} onChange={setSymbol}>
                {snapshot.instruments
                  .filter((i) => i.market === "B3")
                  .map((i) => (
                    <option key={i.symbol}>{i.symbol}</option>
                  ))}
              </Select>
            </Field>
            <Field label="Arquivo (até 1 MiB)">
              <Input
                type="file"
                accept=".csv,text/csv"
                onChange={async (e) => {
                  const f = e.target.files?.[0];
                  if (f) {
                    if (f.size > 1024 * 1024) {
                      setError("Limite de upload: 1 MiB");
                      return;
                    }
                    setCsv(await f.text());
                  }
                }}
              />
            </Field>
          </div>
          <Field label="Fonte / fornecedor">
            <Input
              required
              maxLength={120}
              value={source}
              onChange={(e) => setSource(e.target.value)}
              placeholder="Fornecedor do histórico"
            />
          </Field>
          <Field label="Licença / autorização de uso">
            <Input
              required
              maxLength={300}
              value={license}
              onChange={(e) => setLicense(e.target.value)}
              placeholder="Identifique o contrato ou permissão de uso"
            />
          </Field>
          <Field label="Conteúdo CSV">
            <Textarea
              required
              rows={4}
              value={csv}
              onChange={(e) => setCsv(e.target.value)}
              className="font-mono text-xs"
              placeholder="time,open,high,low,close,volume"
            />
          </Field>
          <label className="flex items-start gap-2 text-xs leading-relaxed text-muted-foreground">
            <input type="checkbox" required className="mt-0.5" />
            Confirmo autorização para armazenar e processar este histórico. Não estou apresentando
            dados sintéticos como mercado real.
          </label>
          <Button disabled={busy}>
            {busy ? <Loader2 className="animate-spin" /> : <Upload />}Importar e validar
          </Button>
          {success && (
            <p role="status" className="text-xs text-emerald-500">
              {success}
            </p>
          )}
        </form>
      </CardContent>
    </Card>
  );
}
function FetchForm({
  configured,
  act,
  busy,
}: {
  configured: boolean;
  act: (p: string, b: unknown) => Promise<void>;
  busy: boolean;
}) {
  const [symbol, setSymbol] = useState("AAPL");
  const [days, setDays] = useState(7);
  return (
    <form
      className="grid gap-4"
      onSubmit={(e) => {
        e.preventDefault();
        void act("/api/datasets/fetch", { symbol, days });
      }}
    >
      <Field label="Ativo">
        <Select value={symbol} onChange={setSymbol}>
          {["AAPL", "NVDA", "TSLA"].map((s) => (
            <option key={s}>{s}</option>
          ))}
        </Select>
      </Field>
      <Field label="Período em dias (1–30)">
        <Input
          type="number"
          min={1}
          max={30}
          value={days}
          onChange={(e) => setDays(Number(e.target.value))}
        />
      </Field>
      <Button disabled={!configured || busy}>
        <RefreshCw />
        Buscar candles IEX
      </Button>
      {!configured && (
        <p className="text-xs text-amber-500">Alpaca Paper ainda não configurada no servidor.</p>
      )}
    </form>
  );
}
function StrategyForm({ snapshot, refresh }: { snapshot: Snapshot; refresh: () => Promise<void> }) {
  const [s, setS] = useState<Strategy>({
    id: "",
    name: "",
    kind: "ma_cross",
    symbol: "PETR4",
    dataset_id: "",
    account_id: "b3",
    version: 1,
    fast: 9,
    slow: 21,
    threshold: 2,
    stop_bps: 100,
    target_bps: 200,
    enabled: false,
  });
  const [status, setStatus] = useState("");
  const [busy, setBusy] = useState(false);
  const change = (key: keyof Strategy, value: string | number) => setS({ ...s, [key]: value });
  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    try {
      await api("/api/strategies", s);
      await refresh();
      setStatus("Configuração criada e desabilitada.");
    } catch (e) {
      setStatus(String(e));
    } finally {
      setBusy(false);
    }
  }
  return (
    <form className="grid gap-4" onSubmit={submit}>
      <Field label="Nome">
        <Input
          required
          minLength={3}
          maxLength={80}
          value={s.name}
          onChange={(e) => change("name", e.target.value)}
        />
      </Field>
      <Field label="Família">
        <Select value={s.kind} onChange={(v) => change("kind", v)}>
          {Object.entries(kindName).map(([k, v]) => (
            <option key={k} value={k}>
              {v}
            </option>
          ))}
        </Select>
      </Field>
      <div className="grid grid-cols-2 gap-3">
        <Field label="Conta">
          <Select
            value={s.account_id}
            onChange={(v) =>
              setS({ ...s, account_id: v, symbol: v === "b3" ? "PETR4" : "AAPL", dataset_id: "" })
            }
          >
            <option value="b3">B3 replay</option>
            <option value="us">Alpaca Paper</option>
          </Select>
        </Field>
        <Field label="Ativo">
          <Select value={s.symbol} onChange={(v) => setS({ ...s, symbol: v, dataset_id: "" })}>
            {snapshot.instruments
              .filter((i) => (i.market === "B3") === (s.account_id === "b3"))
              .map((i) => (
                <option key={i.symbol}>{i.symbol}</option>
              ))}
          </Select>
        </Field>
      </div>
      {s.account_id === "b3" && (
        <Field label="Dataset para replay">
          <DatasetSelect
            snapshot={{
              ...snapshot,
              datasets: snapshot.datasets.filter((d) => d.symbol === s.symbol),
            }}
            value={s.dataset_id}
            onChange={(v) => change("dataset_id", v)}
          />
        </Field>
      )}
      <div className="grid grid-cols-2 gap-3">
        {(
          [
            ["fast", "Janela rápida", 2, 499],
            ["slow", "Janela lenta / lookback", 3, 500],
            ["threshold", "Z-score", 0.1, 10],
            ["stop_bps", "Stop (bps)", 10, 2000],
            ["target_bps", "Alvo (bps)", 10, 5000],
          ] as const
        ).map(([key, label, min, max]) => (
          <Field key={key} label={label}>
            <Input
              type="number"
              step={key === "threshold" ? 0.1 : 1}
              min={min}
              max={max}
              required
              value={s[key]}
              onChange={(e) => change(key, Number(e.target.value))}
            />
          </Field>
        ))}
      </div>
      <p className="text-xs text-muted-foreground">
        100 bps = 1%. Sinais somente em candles encerrados; versão imutável.
      </p>
      <Button disabled={busy}>
        {busy ? <Loader2 className="animate-spin" /> : <Plus />}Criar configuração
      </Button>
      {status && (
        <p role="status" className="text-xs">
          {status}
        </p>
      )}
    </form>
  );
}
function BacktestForm({
  snapshot,
  act,
  busy,
}: {
  snapshot: Snapshot;
  act: (p: string, b: unknown) => Promise<void>;
  busy: boolean;
}) {
  const [strategy, setStrategy] = useState("");
  const [dataset, setDataset] = useState("");
  const [partition, setPartition] = useState("test");
  const s = snapshot.strategies.find((s) => s.id === strategy)?.payload;
  return (
    <form
      className="grid gap-4 md:grid-cols-4"
      onSubmit={(e) => {
        e.preventDefault();
        void act("/api/backtests", { strategy_id: strategy, dataset_id: dataset, partition });
      }}
    >
      <Field label="Estratégia">
        <Select
          value={strategy}
          onChange={(v) => {
            setStrategy(v);
            setDataset("");
          }}
        >
          <option value="">Selecione</option>
          {snapshot.strategies.map((s) => (
            <option key={s.id} value={s.id}>
              {s.payload.name} · {s.payload.symbol}
            </option>
          ))}
        </Select>
      </Field>
      <Field label="Dataset">
        <DatasetSelect
          snapshot={{
            ...snapshot,
            datasets: snapshot.datasets.filter((d) => d.symbol === s?.symbol),
          }}
          value={dataset}
          onChange={setDataset}
        />
      </Field>
      <Field label="Partição temporal">
        <Select value={partition} onChange={setPartition}>
          <option value="test">Teste final · últimos 30%</option>
          <option value="train">Desenvolvimento · primeiros 70%</option>
          <option value="all">Exploratório · todo período</option>
        </Select>
      </Field>
      <div className="flex items-end">
        <Button className="w-full" disabled={!strategy || !dataset || busy}>
          <FlaskConical />
          Enfileirar backtest
        </Button>
      </div>
    </form>
  );
}
function RiskForm({
  account,
  act,
  busy,
}: {
  account: Account;
  act: (p: string, b: unknown) => Promise<void>;
  busy: boolean;
}) {
  const [p, setP] = useState<Policy>(account.policy);
  return (
    <form
      className="grid gap-4 sm:grid-cols-2"
      onSubmit={(e) => {
        e.preventDefault();
        void act(`/api/accounts/${account.id}/policy`, p);
      }}
    >
      {(
        [
          ["risk_bps", "Risco por operação (bps)", 1, 100],
          ["daily_loss_bps", "Perda diária máxima (bps)", 1, 500],
          ["exposure_bps", "Exposição máxima (bps)", 100, 10000],
          ["max_positions", "Máximo de posições", 1, 6],
          ["cost_bps", "Custos simulados por lado (bps)", 0, 100],
          ["slippage_bps", "Slippage / limite (bps)", 1, 100],
          ["max_age_seconds", "Idade máxima do preço (segundos)", 60, 900],
        ] as const
      ).map(([key, label, min, max]) => (
        <Field key={key} label={label}>
          <Input
            disabled={!account.halted}
            type="number"
            min={min}
            max={max}
            step={1}
            required
            value={p[key]}
            onChange={(e) => setP({ ...p, [key]: Number(e.target.value) })}
          />
        </Field>
      ))}
      <p className="text-xs leading-relaxed text-muted-foreground sm:col-span-2">
        50 bps = 0,5%. A perda diária considera patrimônio marcado e custos. Limites não são
        garantias de perda máxima: gaps e indisponibilidade podem excedê-los.
      </p>
      <Button disabled={!account.halted || busy} className="sm:col-span-2">
        Salvar política e registrar auditoria
      </Button>
    </form>
  );
}
