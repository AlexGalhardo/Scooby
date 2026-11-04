package trading

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

// FuzzImportCSV feeds arbitrary text to the CSV importer: it must never panic, and whatever it accepts
// must be a valid, strictly ordered series.
func FuzzImportCSV(f *testing.F) {
	f.Add(SampleCSV(), uint8(0))
	f.Add(cryptoCSV(8), uint8(1))
	f.Add("", uint8(0))
	f.Add("time,open,high,low,close,volume\n", uint8(2))
	f.Add("time,open,high,low,close,volume\n2026-10-26T10:00:00-03:00,NaN,30,29,30,100\n", uint8(0))
	f.Add("time,open,high,low,close,volume\n2026-10-26T10:00:00-03:00,1e999,30,29,30,100\n", uint8(0))
	symbols := []string{"PETR4", "BTCUSDT", "EURUSD", "AAPL"}
	f.Fuzz(func(t *testing.T, input string, pick uint8) {
		ds, err := ImportCSV(strings.NewReader(input), symbols[int(pick)%len(symbols)], "fuzz", "fuzz")
		if err != nil {
			return
		}
		for i, c := range ds.Candles {
			if c.Validate() != nil {
				t.Fatalf("accepted an invalid candle at %d", i)
			}
			if i > 0 && !c.Time.After(ds.Candles[i-1].Time) {
				t.Fatalf("accepted an unordered series at %d", i)
			}
		}
	})
}

// FuzzSignal runs every strategy kind over generated candles with fuzzed parameters. Parameters that
// validate must never make a signal panic or return anything but buy, sell or hold.
func FuzzSignal(f *testing.F) {
	f.Add(uint8(0), uint8(3), uint8(9), 2.0, int64(1))
	f.Add(uint8(3), uint8(6), uint8(12), 0.5, int64(7))
	f.Add(uint8(5), uint8(14), uint8(40), 25.0, int64(42))
	f.Add(uint8(6), uint8(4), uint8(20), 1.0, int64(99))
	kinds := Kinds()
	f.Fuzz(func(t *testing.T, kind, fast, slow uint8, threshold float64, seed int64) {
		spec := kinds[int(kind)%len(kinds)]
		s := Strategy{
			Kind: spec.Kind, Symbol: "AAPL", Fast: int(fast), Slow: int(slow), Threshold: threshold,
			StopBPS: 100, TargetBPS: 200,
		}
		if s.Validate() != nil {
			return
		}
		cs := make([]Candle, 0, 140)
		price := 100.0
		start := time.Date(2026, 10, 26, 14, 0, 0, 0, time.UTC)
		x := uint64(seed)
		for i := range 140 {
			x = x*6364136223846793005 + 1442695040888963407
			move := (float64(x>>40)/float64(1<<24) - 0.5) * 2
			next := math.Max(1, price+move)
			cs = append(cs, Candle{
				Time: start.Add(time.Duration(i) * 5 * time.Minute), Open: D(fmt.Sprintf("%.2f", price)),
				High:   D(fmt.Sprintf("%.2f", math.Max(price, next)+0.3)),
				Low:    D(fmt.Sprintf("%.2f", math.Max(0.01, math.Min(price, next)-0.3))),
				Close:  D(fmt.Sprintf("%.2f", next)),
				Volume: D("1000"),
			})
			price = next
			switch Signal(s, cs) {
			case "buy", "sell", "hold":
			default:
				t.Fatalf("%s returned an unknown signal", spec.Kind)
			}
		}
	})
}
