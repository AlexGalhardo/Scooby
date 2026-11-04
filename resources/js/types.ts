import type {
  Window as BacktestWindow,
  Candle,
  Coverage,
  Decimal,
  Fill,
  Instrument,
  KindSpec,
  Metrics,
  MonteCarlo,
  Order,
  OrderTemplate,
  ParamSpec,
  Point,
  Policy,
  Position,
  Preferences,
  PriceAlert,
  Result,
  Robustness,
  Strategy,
} from "./gen/wire";

export type {
  BacktestWindow,
  Candle,
  Coverage,
  Decimal,
  Fill,
  Instrument,
  KindSpec,
  Metrics,
  MonteCarlo,
  Order,
  OrderTemplate,
  ParamSpec,
  Point,
  Policy,
  Position,
  Preferences,
  PriceAlert,
  Result,
  Robustness,
  Strategy,
};

export interface User {
  id: number;
  name: string;
  email: string;
}
export interface SharedProps {
  appName: string;
  auth: { user: User | null };
  [key: string]: unknown;
}
export interface Dataset {
  id: string;
  symbol: string;
  currency: string;
  source: string;
  license: string;
  hash: string;
  bars: number;
  first: string;
  last: string;
  last_close: Decimal;
  problems: string[];
  candles?: Candle[];
}
export interface Account {
  id: string;
  mode: string;
  currency: string;
  cash: Decimal;
  equity: Decimal;
  realized: Decimal;
  unrealized: Decimal;
  exposure: Decimal;
  drawdown: Decimal;
  halted: boolean;
  reason: string;
  heartbeat: string;
  reconciled_at: string;
  orders: Order[];
  positions: Record<string, Position>;
  fills: Fill[];
  policy: Policy;
  cursors: Record<string, string>;
  day_start: Decimal;
  peak: Decimal;
}
export interface Job {
  id: string;
  dataset_id: string;
  strategy: Strategy;
  partition: string;
  result?: Result;
  error?: string;
  created_at: string;
  finished_at: string;
}
export interface RecordRow<T> {
  id: string;
  status: string;
  version: number;
  updated_at: string;
  payload: T;
}
export interface Snapshot {
  accounts: Account[];
  coverage: Coverage[];
  datasets: Dataset[];
  strategies: RecordRow<Strategy>[];
  backtests: RecordRow<Job>[];
  audit: {
    id: number;
    actor: string;
    action: string;
    account: string;
    details: string;
    time: string;
  }[];
  instruments: Instrument[];
  settings: {
    google_configured: boolean;
    owner_configured: boolean;
    alpaca_configured: boolean;
    live_trading: boolean;
    mode: string;
    feed: string;
    timeframe: number;
    engine_version: string;
  };
  server_time: string;
}
export const money = (value: string | number, currency = "USD") =>
  /^[A-Z]{3}$/.test(currency)
    ? new Intl.NumberFormat("en-US", { style: "currency", currency }).format(Number(value))
    : `${new Intl.NumberFormat("en-US", { maximumFractionDigits: 2 }).format(Number(value))} ${currency}`;
export const date = (value: string) =>
  !value || value.startsWith("0001") ? "No update" : new Date(value).toLocaleString("en-US");
export const alive = (value: string) =>
  value && !value.startsWith("0001") && Date.now() - new Date(value).getTime() < 45000;
