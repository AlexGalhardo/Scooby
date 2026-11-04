package trading

import (
	"context"
	"testing"
	"time"
)

// BenchmarkReserve measures one risk decision: sizing, limits and reservation for an entry.
func BenchmarkReserve(b *testing.B) {
	s := strategy()
	at := time.Now()
	for i := 0; b.Loop(); i++ {
		a := armed()
		if _, err := a.Reserve(s, "buy", "bench", D("100"), at); err != nil {
			b.Fatal(err)
		}
		_ = i
	}
}

// BenchmarkSignal measures one signal evaluation per kind over a session-sized window.
func BenchmarkSignal(b *testing.B) {
	cs := syntheticDataset(2).Candles
	for _, spec := range Kinds() {
		s := Strategy{Kind: spec.Kind, Symbol: "AAPL"}.Defaults()
		b.Run(spec.Kind, func(b *testing.B) {
			for b.Loop() {
				_ = Signal(s, cs)
			}
		})
	}
}

// BenchmarkBacktest measures a full backtest, with robustness checks, by sample size.
func BenchmarkBacktest(b *testing.B) {
	for _, tc := range []struct {
		name     string
		sessions int
	}{{"5 sessions", 5}, {"20 sessions", 20}, {"60 sessions", 60}} {
		ds := syntheticDataset(tc.sessions)
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				if _, err := Backtest(context.Background(), ds, fastStrategy(), DefaultPolicy(), "all"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
