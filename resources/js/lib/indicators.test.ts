import { expect, test } from "bun:test";
import type { Candle } from "@/types";
import { bollinger, ema, rsi, sma, vwap } from "./indicators";

const candle = (i: number, close: number, volume = 10): Candle => ({
  time: new Date(Date.UTC(2026, 9, 26, 13, i * 5)).toISOString(),
  open: String(close),
  high: String(close + 1),
  low: String(close - 1),
  close: String(close),
  volume: String(volume),
});
const series = [1, 2, 3, 4, 5, 6].map((c, i) => candle(i, c));

test("sma averages the trailing window and starts after warm-up", () => {
  expect(sma(series, 3).map((p) => p.value)).toEqual([2, 3, 4, 5]);
  expect(sma(series, 7)).toEqual([]);
});

test("ema seeds with the sma and weights recent closes", () => {
  const out = ema(series, 3);
  expect(out[0].value).toBe(2);
  expect(out[1].value).toBeCloseTo(3, 10);
  expect(ema(series, 0)).toEqual([]);
});

test("bollinger bands collapse on a flat series and straddle the mean otherwise", () => {
  const flat = [5, 5, 5, 5].map((c, i) => candle(i, c));
  expect(bollinger(flat, 3)[0]).toMatchObject({ upper: 5, middle: 5, lower: 5 });
  const [p] = bollinger(series, 3);
  expect(p.upper).toBeGreaterThan(p.middle);
  expect(p.lower).toBeLessThan(p.middle);
});

test("rsi is 100 on a rising series, 0 on a falling one and 50 on a flat one", () => {
  const up = Array.from({ length: 20 }, (_, i) => candle(i, 10 + i));
  const down = Array.from({ length: 20 }, (_, i) => candle(i, 30 - i));
  const flat = Array.from({ length: 20 }, (_, i) => candle(i, 10));
  expect(rsi(up, 14).at(-1)?.value).toBe(100);
  expect(rsi(down, 14).at(-1)?.value).toBe(0);
  expect(rsi(flat, 14).at(-1)?.value).toBe(50);
  expect(rsi(up.slice(0, 14), 14)).toEqual([]);
});

test("vwap restarts on a new day and skips zero volume", () => {
  const next = { ...candle(0, 100, 5), time: "2026-10-27T13:00:00.000Z" };
  const out = vwap([candle(0, 10), next, candle(1, 20, 0)]);
  expect(out[0].value).toBeCloseTo(10, 10);
  expect(out[1].value).toBeCloseTo(100, 10);
});
