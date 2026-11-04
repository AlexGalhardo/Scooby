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
export type Decimal = string;
export interface Candle {
  time: string;
  open: Decimal;
  high: Decimal;
  low: Decimal;
  close: Decimal;
  volume: Decimal;
}
export interface Instrument {
  symbol: string;
  market: string;
  currency: string;
  step: Decimal;
  tick: Decimal;
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
  problems: string[];
  candles?: Candle[];
}
export interface Strategy {
  id: string;
  name: string;
  kind: string;
  symbol: string;
  dataset_id: string;
  account_id: string;
  version: number;
  fast: number;
  slow: number;
  threshold: number;
  stop_bps: number;
  target_bps: number;
  enabled: boolean;
}
export interface Policy {
  risk_bps: number;
  daily_loss_bps: number;
  exposure_bps: number;
  max_positions: number;
  cost_bps: number;
  slippage_bps: number;
  max_age_seconds: number;
}
export interface Order {
  id: string;
  symbol: string;
  side: string;
  qty: Decimal;
  filled: Decimal;
  limit: Decimal;
  average: Decimal;
  status: string;
  reason: string;
  created_at: string;
}
export interface Position {
  symbol: string;
  qty: Decimal;
  average: Decimal;
  mark: Decimal;
  stop: Decimal;
  target: Decimal;
}
export interface Fill {
  id: string;
  order_id: string;
  symbol: string;
  side: string;
  qty: Decimal;
  price: Decimal;
  fees: Decimal;
  time: string;
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
}
export interface Point {
  time: string;
  equity: Decimal;
  benchmark: Decimal;
}
export interface Result {
  metrics: {
    return_pct: Decimal;
    max_drawdown_pct: Decimal;
    realized: Decimal;
    costs: Decimal;
    fills: number;
    trades: number;
    win_rate: number;
    exposure_bars: number;
  };
  curve: Point[] | null;
  fills: Fill[] | null;
  dataset_hash: string;
  strategy: Strategy;
  policy: Policy;
  warnings: string[];
  engine_version: string;
  partition: string;
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
    mode: string;
    feed: string;
    timeframe: number;
    engine_version: string;
  };
  server_time: string;
}
export const money = (value: string | number, currency = "USD") =>
  new Intl.NumberFormat("en-US", { style: "currency", currency }).format(Number(value));
export const date = (value: string) =>
  !value || value.startsWith("0001") ? "No update" : new Date(value).toLocaleString("en-US");
export const alive = (value: string) =>
  value && !value.startsWith("0001") && Date.now() - new Date(value).getTime() < 45000;
