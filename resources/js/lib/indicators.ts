import type { Candle } from "@/types";

export interface IndicatorPoint {
  time: string;
  value: number;
}

const closes = (candles: Candle[]): number[] => candles.map((c) => Number(c.close));

export function sma(candles: Candle[], period: number): IndicatorPoint[] {
  const out: IndicatorPoint[] = [];
  if (period < 1) return out;
  const values = closes(candles);
  let sum = 0;
  for (let i = 0; i < values.length; i++) {
    sum += values[i];
    if (i >= period) sum -= values[i - period];
    if (i >= period - 1) out.push({ time: candles[i].time, value: sum / period });
  }
  return out;
}

export function ema(candles: Candle[], period: number): IndicatorPoint[] {
  const out: IndicatorPoint[] = [];
  if (period < 1 || candles.length < period) return out;
  const values = closes(candles);
  const k = 2 / (period + 1);
  let prev = values.slice(0, period).reduce((a, b) => a + b, 0) / period;
  out.push({ time: candles[period - 1].time, value: prev });
  for (let i = period; i < values.length; i++) {
    prev = values[i] * k + prev * (1 - k);
    out.push({ time: candles[i].time, value: prev });
  }
  return out;
}

export interface BollingerPoint {
  time: string;
  upper: number;
  middle: number;
  lower: number;
}

export function bollinger(candles: Candle[], period = 20, deviations = 2): BollingerPoint[] {
  const out: BollingerPoint[] = [];
  if (period < 2) return out;
  const values = closes(candles);
  for (let i = period - 1; i < values.length; i++) {
    const window = values.slice(i - period + 1, i + 1);
    const mean = window.reduce((a, b) => a + b, 0) / period;
    const variance = window.reduce((a, b) => a + (b - mean) ** 2, 0) / period;
    const spread = Math.sqrt(variance) * deviations;
    out.push({ time: candles[i].time, upper: mean + spread, middle: mean, lower: mean - spread });
  }
  return out;
}

// Wilder smoothing; the first value uses the plain average of the first `period` changes.
export function rsi(candles: Candle[], period = 14): IndicatorPoint[] {
  const out: IndicatorPoint[] = [];
  if (period < 1 || candles.length <= period) return out;
  const values = closes(candles);
  let gain = 0;
  let loss = 0;
  for (let i = 1; i <= period; i++) {
    const change = values[i] - values[i - 1];
    if (change >= 0) gain += change;
    else loss -= change;
  }
  gain /= period;
  loss /= period;
  const toRSI = () => (loss === 0 ? (gain === 0 ? 50 : 100) : 100 - 100 / (1 + gain / loss));
  out.push({ time: candles[period].time, value: toRSI() });
  for (let i = period + 1; i < values.length; i++) {
    const change = values[i] - values[i - 1];
    gain = (gain * (period - 1) + Math.max(change, 0)) / period;
    loss = (loss * (period - 1) + Math.max(-change, 0)) / period;
    out.push({ time: candles[i].time, value: toRSI() });
  }
  return out;
}

// Session VWAP: the cumulative sum restarts on each UTC calendar day.
export function vwap(candles: Candle[]): IndicatorPoint[] {
  const out: IndicatorPoint[] = [];
  let day = "";
  let pv = 0;
  let volume = 0;
  for (const c of candles) {
    const d = c.time.slice(0, 10);
    if (d !== day) {
      day = d;
      pv = 0;
      volume = 0;
    }
    const v = Number(c.volume);
    pv += ((Number(c.high) + Number(c.low) + Number(c.close)) / 3) * v;
    volume += v;
    if (volume > 0) out.push({ time: c.time, value: pv / volume });
  }
  return out;
}
