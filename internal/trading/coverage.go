package trading

// Coverage describes what the terminal can do for one market: where its data comes from, how fresh it
// is, and what can execute. It is shown in the web terminal so no market implies more than it has.
type Coverage struct {
	Market    string `json:"market"`
	Label     string `json:"label"`
	Data      string `json:"data"`      // imported, delayed, history or none
	Execution string `json:"execution"` // paper_broker, replay or backtest_only
	Source    string `json:"source"`
	Notes     string `json:"notes"`
}

// CoverageMatrix lists every market in display order.
func CoverageMatrix() []Coverage {
	return []Coverage{
		{
			Market:    "B3",
			Label:     "B3 equities",
			Data:      "imported",
			Execution: "replay",
			Source:    "Licensed CSV import (5-minute candles)",
			Notes:     "No live feed. Replay is not a broker. The calendar is reviewed for 2026 only and unknown years fail closed.",
		},
		{
			Market:    "US",
			Label:     "US equities and ETFs",
			Data:      "delayed",
			Execution: "paper_broker",
			Source:    "Alpaca IEX history, about 16 minutes behind",
			Notes:     "IEX is not the consolidated market. Orders go to Alpaca Paper only; live hosts cannot be configured.",
		},
		{
			Market: "CRYPTO", Label: "Crypto spot", Data: "history", Execution: "backtest_only",
			Source: "Binance public klines (closed 5-minute candles) or CSV import",
			Notes:  "24/7 sessions in UTC. No paper adapter yet; backtests and research only.",
		},
		{
			Market: "FX", Label: "Spot forex", Data: "imported", Execution: "backtest_only",
			Source: "CSV import (5-minute candles)",
			Notes:  "No central tape and spreads vary: results are indicative. Sessions run Monday to Friday in UTC.",
		},
	}
}
