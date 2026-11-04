# OpenStock: patterns for vukem-trading

Source: [Open-Dev-Society/OpenStock](https://github.com/Open-Dev-Society/OpenStock) README. Researched 2026-11-04. License **AGPL-3.0**: do not copy code into this repository; use it only as a product reference.

## What it is

A Robinhood-style web dashboard for eight markets (US, India, Germany, Canada, Australia, Saudi Arabia, crypto, forex). Stack: Next.js 15, React 19, TypeScript, Tailwind v4, shadcn/ui, Better Auth with MongoDB, Inngest for background jobs. Quotes, search, profiles and news come from Finnhub (free key); charts and financials are TradingView widgets. Not a brokerage; data may be delayed.

## Features

- Command-palette symbol search, sector heatmap, market leaders and news.
- Candles with eight indicators (SMA, EMA, Bollinger, VWAP, RSI, MACD, Stochastic, ATR).
- Watchlists capped at 50 symbols.
- Price alerts above or below a target, checked every five minutes and sent by email.
- Weekly AI news digest. No screener; AI is not used for stock analysis.

## Coverage matrix

The README publishes, per market, what is available: live charts and quotes (US, crypto), charts only (forex), delayed charts (Canada, Australia), end-of-day charts (India, Germany), blocked (Saudi). Showing this explicitly avoids implying live data where there is none.

## Pitfall

`NEXT_PUBLIC_OPENSTOCK_DATA_MODE` (cached versus realtime) is a build-time variable, so changing it needs a rebuild. Prefer a runtime setting.

## Recommendations for vukem-trading

1. Show a per-market coverage matrix (live, delayed, end of day, blocked) in the UI and in the docs.
2. Model alerts as threshold plus checker interval plus delivery channel; the project already has optional Telegram alerts.
3. Cap watchlist size to keep polling costs bounded.
4. Keep data mode a runtime setting.
