import { useEffect, useRef } from "react";
import type { Candle } from "@/types";

const rows = 18;

/**
 * A time-and-sales tape and a volume-at-price ladder, drawn on canvas. The data is handed to the
 * canvas in an effect and never kept in React state, so a faster feed repaints without re-rendering
 * the page.
 */
export function TapeLadder({ candles }: { candles: Candle[] }) {
  const tape = useRef<HTMLCanvasElement>(null);
  const ladder = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const recent = candles.slice(-rows);
    const tapeCtx = tape.current?.getContext("2d");
    if (tape.current && tapeCtx) {
      const { width, height } = tape.current;
      tapeCtx.clearRect(0, 0, width, height);
      tapeCtx.font = "12px ui-monospace, monospace";
      recent.forEach((c, i) => {
        const up = Number(c.close) >= Number(c.open);
        tapeCtx.fillStyle = up ? "#10b981" : "#ef4444";
        const time = new Date(c.time).toLocaleTimeString("en-US", { hour12: false });
        tapeCtx.fillText(
          `${time}  ${Number(c.close).toFixed(2).padStart(10)}  ${Number(c.volume).toFixed(0).padStart(9)}`,
          8,
          16 + i * 18,
        );
      });
    }
    const ladderCtx = ladder.current?.getContext("2d");
    if (ladder.current && ladderCtx && recent.length > 0) {
      const { width, height } = ladder.current;
      ladderCtx.clearRect(0, 0, width, height);
      const lows = recent.map((c) => Number(c.low));
      const highs = recent.map((c) => Number(c.high));
      const lo = Math.min(...lows);
      const hi = Math.max(...highs);
      const step = (hi - lo) / rows || 1;
      const volume = new Array<number>(rows).fill(0);
      for (const c of recent) {
        const bucket = Math.min(rows - 1, Math.floor((Number(c.close) - lo) / step));
        volume[bucket] += Number(c.volume);
      }
      const peak = Math.max(...volume, 1);
      ladderCtx.font = "11px ui-monospace, monospace";
      volume.forEach((v, i) => {
        const y = (rows - 1 - i) * 18;
        const price = lo + step * i;
        ladderCtx.fillStyle = "rgba(16,185,129,0.35)";
        ladderCtx.fillRect(86, y + 2, ((width - 90) * v) / peak, 14);
        ladderCtx.fillStyle = "#94a3b8";
        ladderCtx.fillText(price.toFixed(2), 4, y + 13);
        if (height < y) return;
      });
    }
  }, [candles]);

  const last = candles[candles.length - 1];
  return (
    <div className="grid gap-4 sm:grid-cols-2">
      <canvas
        ref={tape}
        width={420}
        height={rows * 18 + 8}
        className="w-full rounded-lg border bg-muted/20"
        role="img"
        aria-label={`Time and sales tape, latest close ${last ? last.close : "unavailable"}`}
      />
      <canvas
        ref={ladder}
        width={420}
        height={rows * 18 + 8}
        className="w-full rounded-lg border bg-muted/20"
        role="img"
        aria-label="Volume at price ladder for the latest candles"
      />
    </div>
  );
}
