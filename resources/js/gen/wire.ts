// Code generated from internal/trading by `go test ./internal/trading -run Wire -update`. DO NOT EDIT.

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
  draft: boolean;
}

export interface Policy {
  risk_bps: number;
  daily_loss_bps: number;
  exposure_bps: number;
  max_positions: number;
  cost_bps: number;
  slippage_bps: number;
  max_age_seconds: number;
  stop_streak: number;
  streak_lock_mins: number;
  cooldown_mins: number;
  drawdown_bps: number;
}

export interface Order {
  id: string;
  key: string;
  symbol: string;
  side: string;
  qty: Decimal;
  filled: Decimal;
  average: Decimal;
  limit: Decimal;
  stop: Decimal;
  target: Decimal;
  reserved: Decimal;
  status: string;
  broker_id: string;
  reason: string;
  close_reason: string;
  stop_bps: number;
  target_bps: number;
  generation: number;
  created_at: string;
  updated_at: string;
}

export interface Position {
  symbol: string;
  qty: Decimal;
  average: Decimal;
  mark: Decimal;
  stop: Decimal;
  target: Decimal;
  tranches: number;
  last_entry: Decimal;
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

export interface Point {
  time: string;
  equity: Decimal;
  benchmark: Decimal;
}

export interface Metrics {
  benchmark_return_pct: Decimal;
  turnover_pct: Decimal;
  average_win: Decimal;
  average_loss: Decimal;
  max_drawdown_seconds: number;
  sharpe: number;
  sortino: number;
  calmar: number;
  sqn: number;
  annual_return_pct: number;
  exposure_pct: number;
  risk_free_annual_pct: number;
  periods_per_year: number;
  return_pct: Decimal;
  max_drawdown_pct: Decimal;
  realized: Decimal;
  costs: Decimal;
  fills: number;
  trades: number;
  win_rate: number;
  exposure_bars: number;
}

export interface Window {
  index: number;
  bars: number;
  trades: number;
  return_pct: Decimal;
  max_drawdown_pct: Decimal;
  enough_trades: boolean;
  note?: string;
}

export interface MonteCarlo {
  runs: number;
  block: number;
  p5_return_pct: number;
  p50_return_pct: number;
  p95_return_pct: number;
  probability_of_loss: number;
  trades: number;
  seed: number;
  technique: string;
}

export interface Robustness {
  trials: number;
  low_trades: boolean;
  deflated_sharpe: number;
  monte_carlo?: MonteCarlo;
  cost_stress_return_pct?: Decimal;
  cost_stress_note?: string;
}

export interface Result {
  calendar_source: string;
  metrics: Metrics;
  curve: Point[] | null;
  fills: Fill[] | null;
  dataset_hash: string;
  strategy: Strategy;
  policy: Policy;
  warnings: string[] | null;
  engine_version: string;
  partition: string;
  trade_pnl: Decimal[] | null;
  windows?: Window[];
  robustness?: Robustness;
}

export interface ParamSpec {
  name: string;
  label: string;
  type: string;
  min: number;
  max: number;
  step: number;
  default: number;
}

export interface KindSpec {
  kind: string;
  label: string;
  description: string;
  params: ParamSpec[] | null;
  scales: boolean;
}

export interface Coverage {
  market: string;
  label: string;
  data: string;
  execution: string;
  source: string;
  notes: string;
}

export interface PriceAlert {
  id: string;
  symbol: string;
  direction: string;
  price: Decimal;
  active: boolean;
  triggered_at: string;
}

export interface OrderTemplate {
  key: string;
  label: string;
  side: string;
  cash_pct: number;
  stop_bps: number;
  target_bps: number;
}

export interface Preferences {
  account: string;
  symbol: string;
  watchlist: string[] | null;
  alerts: PriceAlert[] | null;
  templates: OrderTemplate[] | null;
}
