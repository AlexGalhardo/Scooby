import {
  CandlestickSeries,
  ColorType,
  createChart,
  HistogramSeries,
  LineSeries,
  type UTCTimestamp,
} from "lightweight-charts";
import { useEffect, useRef, useState } from "react";
import { bollinger, ema, type IndicatorPoint, rsi, sma, vwap } from "@/lib/indicators";
import type { Candle, Point } from "@/types";

const INDICATORS = [
  { key: "sma", label: "SMA 20" },
  { key: "ema", label: "EMA 50" },
  { key: "bollinger", label: "Bollinger 20" },
  { key: "vwap", label: "VWAP" },
  { key: "volume", label: "Volume" },
  { key: "rsi", label: "RSI 14" },
] as const;
type IndicatorKey = (typeof INDICATORS)[number]["key"];

export default function MarketChart({ candles, curve }: { candles?: Candle[]; curve?: Point[] }) {
  const ref = useRef<HTMLDivElement>(null);
  const [on, setOn] = useState<Set<IndicatorKey>>(new Set());
  useEffect(() => {
    if (!ref.current) return;
    const dark = document.documentElement.classList.contains("dark");
    const chart = createChart(ref.current, {
      height: on.has("rsi") || on.has("volume") ? 420 : 300,
      layout: {
        background: { type: ColorType.Solid, color: "transparent" },
        textColor: dark ? "#9ca3af" : "#6b7280",
        attributionLogo: true,
      },
      grid: {
        vertLines: { color: dark ? "#ffffff06" : "#00000006" },
        horzLines: { color: dark ? "#ffffff08" : "#00000008" },
      },
      rightPriceScale: { borderVisible: false },
      timeScale: { borderVisible: false, timeVisible: true },
      handleScroll: true,
    });
    const timestamp = (time: string) => Math.floor(new Date(time).getTime() / 1000) as UTCTimestamp;
    const line = (points: IndicatorPoint[], color: string, title: string, pane = 0) =>
      chart
        .addSeries(
          LineSeries,
          { color, lineWidth: 1, title, priceLineVisible: false, lastValueVisible: false },
          pane,
        )
        .setData(points.map((p) => ({ time: timestamp(p.time), value: p.value })));
    if (candles) {
      chart
        .addSeries(CandlestickSeries, {
          upColor: "#34d399",
          downColor: "#f87171",
          borderVisible: false,
          wickUpColor: "#34d399",
          wickDownColor: "#f87171",
        })
        .setData(
          candles.map((c) => ({
            time: timestamp(c.time),
            open: Number(c.open),
            high: Number(c.high),
            low: Number(c.low),
            close: Number(c.close),
          })),
        );
    }
    if (candles && on.has("sma")) line(sma(candles, 20), "#60a5fa", "SMA 20");
    if (candles && on.has("ema")) line(ema(candles, 50), "#f59e0b", "EMA 50");
    if (candles && on.has("vwap")) line(vwap(candles), "#c084fc", "VWAP");
    if (candles && on.has("bollinger")) {
      const bands = bollinger(candles, 20, 2);
      line(
        bands.map((b) => ({ time: b.time, value: b.upper })),
        "#94a3b8",
        "BB upper",
      );
      line(
        bands.map((b) => ({ time: b.time, value: b.lower })),
        "#94a3b8",
        "BB lower",
      );
    }
    let pane = 1;
    if (candles && on.has("volume")) {
      chart
        .addSeries(HistogramSeries, { priceFormat: { type: "volume" }, title: "Volume" }, pane++)
        .setData(
          candles.map((c) => ({
            time: timestamp(c.time),
            value: Number(c.volume),
            color: Number(c.close) >= Number(c.open) ? "#34d39988" : "#f8717188",
          })),
        );
    }
    if (candles && on.has("rsi")) line(rsi(candles, 14), "#2dd4bf", "RSI 14", pane);
    if (curve) {
      chart
        .addSeries(LineSeries, { color: "#34d399", lineWidth: 2, title: "Equity" })
        .setData(curve.map((p) => ({ time: timestamp(p.time), value: Number(p.equity) })));
      chart
        .addSeries(LineSeries, { color: "#64748b", lineWidth: 1, title: "Benchmark" })
        .setData(curve.map((p) => ({ time: timestamp(p.time), value: Number(p.benchmark) })));
    }
    chart.timeScale().fitContent();
    const observer = new ResizeObserver((entries) =>
      chart.applyOptions({ width: entries[0].contentRect.width }),
    );
    observer.observe(ref.current);
    return () => {
      observer.disconnect();
      chart.remove();
    };
  }, [candles, curve, on]);
  return (
    <div>
      {candles && (
        <fieldset className="mb-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
          <legend className="sr-only">Chart indicators</legend>
          {INDICATORS.map(({ key, label }) => (
            <label key={key} className="flex items-center gap-1.5">
              <input
                type="checkbox"
                checked={on.has(key)}
                onChange={(e) => {
                  const next = new Set(on);
                  if (e.target.checked) next.add(key);
                  else next.delete(key);
                  setOn(next);
                }}
              />
              {label}
            </label>
          ))}
        </fieldset>
      )}
      <div
        ref={ref}
        role="img"
        className="w-full"
        aria-label="Market chart; detailed values in the tables"
      />
    </div>
  );
}
