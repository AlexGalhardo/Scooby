import { OctagonX, Plus, Trash2 } from "lucide-react";
import { type ReactNode, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  type Account,
  type Coverage,
  type Dataset,
  type Instrument,
  money,
  type Preferences,
  type PriceAlert,
} from "@/types";

function PlainSelect({
  value,
  onChange,
  label,
  children,
}: {
  value: string;
  onChange: (v: string) => void;
  label: string;
  children: ReactNode;
}) {
  return (
    <select
      aria-label={label}
      className="h-9 w-full rounded-lg border bg-background px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-emerald-500"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    >
      {children}
    </select>
  );
}

const bps = (part: number, whole: number) => (whole > 0 ? (part / whole) * 10000 : 0);

/** The safety strip: what the account may do right now and how much of each limit is already used. */
export function SafetyBar({
  account,
  act,
  busy,
}: {
  account: Account;
  act: (path: string, body: unknown) => Promise<void>;
  busy: boolean;
}) {
  const equity = Number(account.equity);
  const lossBps = Math.max(0, bps(Number(account.day_start) - equity, Number(account.day_start)));
  const drawdownBps = bps(Number(account.peak) - equity, Number(account.peak));
  const exposureBps = bps(Number(account.exposure), equity);
  const open = account.orders.filter((o) =>
    ["approved", "dispatching", "unknown", "acknowledged", "partially_filled"].includes(o.status),
  ).length;
  const chips: [string, string, boolean][] = [
    [
      "Daily loss",
      `${lossBps.toFixed(0)} / ${account.policy.daily_loss_bps} bps`,
      lossBps >= account.policy.daily_loss_bps * 0.8,
    ],
    [
      "Drawdown",
      account.policy.drawdown_bps
        ? `${drawdownBps.toFixed(0)} / ${account.policy.drawdown_bps} bps`
        : `${drawdownBps.toFixed(0)} bps`,
      account.policy.drawdown_bps > 0 && drawdownBps >= account.policy.drawdown_bps * 0.8,
    ],
    [
      "Exposure",
      `${exposureBps.toFixed(0)} / ${account.policy.exposure_bps} bps`,
      exposureBps >= account.policy.exposure_bps * 0.9,
    ],
    ["Open orders", String(open), false],
  ];
  return (
    <section
      aria-label="Safety status"
      className={`mb-6 flex flex-wrap items-center gap-3 rounded-xl border px-4 py-3 text-xs ${
        account.halted
          ? "border-amber-500/40 bg-amber-500/5"
          : "border-emerald-500/30 bg-emerald-500/5"
      }`}
    >
      <p role="status" className="font-medium">
        {account.halted
          ? `Blocked · ${account.reason || "entries are disabled"}`
          : "Armed · entries enabled"}
      </p>
      <ul className="flex flex-wrap gap-2">
        {chips.map(([label, value, warn]) => (
          <li
            key={label}
            className={`rounded-md border px-2 py-1 font-mono ${warn ? "border-amber-500/50 text-amber-500" : ""}`}
          >
            {label}: {value}
          </li>
        ))}
      </ul>
      {!account.halted && (
        <Button
          size="sm"
          variant="destructive"
          disabled={busy}
          className="ml-auto"
          onClick={() => void act(`/api/accounts/${account.id}/control`, { action: "halt" })}
        >
          <OctagonX />
          Kill switch
        </Button>
      )}
    </section>
  );
}

/** The saved watchlist with the data mode of each market, and the price alerts. */
export function WatchPanel({
  instruments,
  datasets,
  coverage,
  prefs,
  save,
}: {
  instruments: Instrument[];
  datasets: Dataset[];
  coverage: Coverage[];
  prefs: Preferences;
  save: (next: Preferences) => void;
}) {
  const watchlist = prefs.watchlist ?? [];
  const alerts = prefs.alerts ?? [];
  const [add, setAdd] = useState("");
  const [alert, setAlert] = useState({ symbol: prefs.symbol, direction: "above", price: "" });
  const mode = (symbol: string) => {
    const market = instruments.find((i) => i.symbol === symbol)?.market;
    return coverage.find((c) => c.market === market)?.data ?? "none";
  };
  const price = (symbol: string) => {
    const latest = datasets
      .filter((d) => d.symbol === symbol)
      .sort((a, b) => b.last.localeCompare(a.last))[0];
    const currency = instruments.find((i) => i.symbol === symbol)?.currency ?? "USD";
    return latest ? money(latest.last_close, currency) : "no data";
  };
  const addable = instruments.filter((i) => !watchlist.includes(i.symbol));
  const atCap = watchlist.length >= 20;
  return (
    <Card>
      <CardHeader>
        <CardTitle>Watchlist and alerts</CardTitle>
        <CardDescription>
          Latest close per symbol and the data mode of its market. Alerts are checked by the engine
          on every tick.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-5">
        <ul className="grid gap-2 text-sm">
          {watchlist.map((symbol) => (
            <li
              key={symbol}
              className="flex items-center justify-between gap-3 rounded-md border px-3 py-2"
            >
              <button
                type="button"
                className="font-mono underline-offset-2 hover:underline"
                aria-label={`Select ${symbol} for the ticket`}
                onClick={() => save({ ...prefs, symbol })}
              >
                {symbol}
                {prefs.symbol === symbol ? " ·" : ""}
              </button>
              <span className="text-xs text-muted-foreground">
                {price(symbol)} · {mode(symbol)}
              </span>
              <Button
                size="icon"
                variant="ghost"
                aria-label={`Remove ${symbol}`}
                onClick={() => save({ ...prefs, watchlist: watchlist.filter((s) => s !== symbol) })}
              >
                <Trash2 className="size-3" />
              </Button>
            </li>
          ))}
        </ul>
        <div className="flex items-end gap-2">
          <Label className="grid flex-1 gap-2 text-xs font-normal text-muted-foreground">
            Add a symbol
            <PlainSelect label="Add a symbol" value={add} onChange={setAdd}>
              <option value="">Choose…</option>
              {addable.map((i) => (
                <option key={i.symbol}>{i.symbol}</option>
              ))}
            </PlainSelect>
          </Label>
          <Button
            variant="outline"
            disabled={!add || atCap}
            onClick={() => {
              save({ ...prefs, watchlist: [...watchlist, add] });
              setAdd("");
            }}
          >
            <Plus />
            Add
          </Button>
        </div>
        {atCap && <p className="text-xs text-amber-500">The watchlist holds at most 20 symbols.</p>}
        <form
          className="grid gap-2 sm:grid-cols-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (!alert.price) return;
            const next: PriceAlert = {
              id: crypto.randomUUID(),
              symbol: alert.symbol,
              direction: alert.direction,
              price: alert.price,
              active: true,
              triggered_at: "0001-01-01T00:00:00Z",
            };
            save({ ...prefs, alerts: [...alerts, next] });
            setAlert({ ...alert, price: "" });
          }}
        >
          <PlainSelect
            label="Alert symbol"
            value={alert.symbol}
            onChange={(v) => setAlert({ ...alert, symbol: v })}
          >
            {instruments.map((i) => (
              <option key={i.symbol}>{i.symbol}</option>
            ))}
          </PlainSelect>
          <PlainSelect
            label="Alert direction"
            value={alert.direction}
            onChange={(v) => setAlert({ ...alert, direction: v })}
          >
            <option value="above">goes above</option>
            <option value="below">goes below</option>
          </PlainSelect>
          <Input
            aria-label="Alert price"
            inputMode="decimal"
            placeholder="Price"
            value={alert.price}
            onChange={(e) => setAlert({ ...alert, price: e.target.value })}
          />
          <Button variant="outline" disabled={!alert.price}>
            Set alert
          </Button>
        </form>
        <ul className="grid gap-1 text-xs">
          {alerts.map((a) => (
            <li
              key={a.id}
              className="flex items-center justify-between rounded-md border px-3 py-2"
            >
              <span>
                {a.symbol} {a.direction} {a.price} · {a.active ? "active" : "fired"}
              </span>
              <Button
                size="icon"
                variant="ghost"
                aria-label={`Delete the ${a.symbol} alert`}
                onClick={() => save({ ...prefs, alerts: alerts.filter((x) => x.id !== a.id) })}
              >
                <Trash2 className="size-3" />
              </Button>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  );
}

/** One-key paper orders. Each template runs through the risk engine on the server. */
export function OrderTicket({
  prefs,
  accountId,
  place,
  busy,
}: {
  prefs: Preferences;
  accountId: string;
  place: (key: string) => void;
  busy: boolean;
}) {
  const templates = prefs.templates ?? [];
  const supported = accountId === "b3" || accountId === "us";
  return (
    <Card>
      <CardHeader>
        <CardTitle>Order ticket</CardTitle>
        <CardDescription>
          Symbol {prefs.symbol} on the {accountId.toUpperCase()} account. Press a hotkey outside a
          field, or click. Every order is sized by the risk engine and by the cash share of the
          template.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-wrap gap-2">
        {templates.map((t) => (
          <Button
            key={t.key}
            variant="outline"
            disabled={!supported || busy}
            aria-keyshortcuts={t.key}
            onClick={() => place(t.key)}
          >
            {t.label} <kbd className="ml-2 rounded border px-1 text-[10px]">{t.key}</kbd>
          </Button>
        ))}
        {!supported && (
          <p className="text-xs text-muted-foreground">Research markets only run backtests.</p>
        )}
      </CardContent>
    </Card>
  );
}
