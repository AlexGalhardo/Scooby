export const site = {
  name: "vukem-trading",
  tagline: "A personal quantitative trading terminal for backtesting and paper trading, written in Go.",
  url: process.env.NEXT_PUBLIC_SITE_URL ?? "https://trading.vukem.dev",
  repo: "https://github.com/AlexGalhardo/vukem-trading",
  editBase: "https://github.com/AlexGalhardo/vukem-trading/edit/main/site/content",
} as const;
