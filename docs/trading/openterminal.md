# OpenTerminal: patterns for the vukem-trading web terminal

Source: [ErTasselli/OpenTerminal](https://github.com/ErTasselli/OpenTerminal) README (MIT). Researched 2026-11-04.

## What it is

A self-hosted, keyboard-driven, dark, widget-based market terminal that runs on public data with no keys or subscriptions. Stack: Next.js 15, React 19, Tailwind 4, Zustand, TanStack Query, lightweight-charts, D3, Recharts; a Node and Express backend with SQLite in WAL mode; an optional AI assistant.

## Data sources (public, not licensed feeds)

Nasdaq public API with Yahoo and Stooq fallbacks for US quotes; the TradingView scanner for fundamentals, screener and heatmap; Nasdaq then Yahoo for options; Yahoo and Google News RSS; CoinGecko and Binance for crypto; FRED for yields and VIX; an economic calendar feed. The README warns that provider terms often limit use to personal or research purposes.

## Features

Drag-and-resize widget workspace with saved layout; command palette; charts with eight timeframes and six indicators; quote panel with fundamentals; aggregated news; sortable full-market screener; sector heatmap; options chain; crypto board; macro dashboard; portfolio tracker with a buy and sell log and realized and unrealized P&L; economic and earnings calendar; shortcuts for palette and widgets. No backtesting, no execution and no price alerts.

## Security defaults worth copying

The API binds to loopback and allows only the local browser origin. Portfolio and AI routes need a shared secret; if none is set, one is generated into a data file. The AI route is rate limited. Trusting forwarded headers is opt-in behind a controlling proxy. The README also lists unresolved dependency advisories, a reminder to run audits in CI.

## Recommendations for vukem-trading

1. Persist a widget layout per user and add a command palette with numbered shortcuts.
2. Copy the security defaults: loopback bind, origin allowlist, a secret on write routes, rate limits on AI routes, opt-in proxy trust.
3. Treat public-endpoint data as unlicensed: put B3 and paid feeds behind provider interfaces and record each source's terms in `docs/fontes.md`.
4. Add a portfolio view with realized and unrealized P&L (already partly present).
