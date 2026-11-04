package trading

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func mk(at time.Time, open, high, low, closeP, vol string) Candle {
	return Candle{Time: at, Open: D(open), High: D(high), Low: D(low), Close: D(closeP), Volume: D(vol)}
}

// session returns bars of one US session starting at 10:00 New York time on day offset.
func session(dayOffset int, rows [][5]string) []Candle {
	start := time.Date(2026, 10, 26+dayOffset, 14, 0, 0, 0, time.UTC)
	out := make([]Candle, 0, len(rows))
	for i, r := range rows {
		out = append(out, mk(start.Add(time.Duration(i)*5*time.Minute), r[0], r[1], r[2], r[3], r[4]))
	}
	return out
}

func signalsOf(s Strategy, cs []Candle) []string {
	out := make([]string, len(cs))
	for i := range cs {
		out[i] = Signal(s, cs[:i+1])
	}
	return out
}

func TestOpeningRangeBreakout(t *testing.T) {
	s := Strategy{Kind: "opening_range", Symbol: "AAPL", Fast: 3, Slow: 4, StopBPS: 100, TargetBPS: 200}
	day1 := session(0, [][5]string{
		{"100", "101", "99", "100", "1000"},
		{"100", "101", "99.5", "100.5", "1000"},
		{"100.5", "101", "99", "100", "1000"},      // range ends here: high 101, low 99
		{"100", "100.8", "99.6", "100.4", "1000"},  // inside the range
		{"100.4", "102.5", "100.3", "102", "1000"}, // closes above the high: buy
		{"102", "102.3", "100.9", "101.5", "1000"},
		{"101.5", "101.6", "98", "98.5", "1000"}, // closes below the low: sell
	})
	day2 := session(1, [][5]string{
		{"110", "111", "109", "110", "1000"},
		{"110", "111", "109", "110.5", "1000"},
		{"110.5", "111", "109", "110", "1000"},
		{"110", "111.2", "109.5", "110.5", "1000"},
	})
	cs := append(day1, day2...)
	got := signalsOf(s, cs)
	want := []string{"hold", "hold", "hold", "hold", "buy", "hold", "sell", "hold", "hold", "hold", "hold"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("bar %d: got %s, want %s (all: %v)", i, got[i], want[i], got)
		}
	}
}

func TestVWAPReversion(t *testing.T) {
	s := Strategy{
		Kind:      "vwap_reversion",
		Symbol:    "AAPL",
		Fast:      4,
		Slow:      6,
		Threshold: 0.5,
		StopBPS:   100,
		TargetBPS: 200,
	}
	rows := [][5]string{}
	for range 6 {
		rows = append(rows, [5]string{"100", "100.2", "99.8", "100", "1000"})
	}
	rows = append(rows, [5]string{"100", "100", "98.8", "99", "1000"})
	rows = append(rows, [5]string{"99", "100.6", "99", "100.5", "1000"})
	got := signalsOf(s, session(0, rows))
	if got[2] != "hold" {
		t.Fatalf("too few bars in the session must hold, got %s", got[2])
	}
	if got[6] != "buy" {
		t.Fatalf("a close 1%% below the VWAP must buy, got %s", got[6])
	}
	if got[7] != "sell" {
		t.Fatalf("a close back above the VWAP must exit, got %s", got[7])
	}
}

func trendSeries() []Candle {
	var rows [][5]string
	price := 100.0
	for i := range 40 {
		move := -0.05
		if i >= 20 {
			move = 0.6
		}
		next := price + move
		rows = append(rows, [5]string{
			fmt.Sprintf("%.2f", price), fmt.Sprintf("%.2f", math.Max(price, next)+0.2),
			fmt.Sprintf("%.2f", math.Min(price, next)-0.2), fmt.Sprintf("%.2f", next), "1000",
		})
		price = next
	}
	return session(0, rows)
}

func TestADXTrendNeedsAStrongTrend(t *testing.T) {
	cs := trendSeries()
	count := func(threshold float64) (buys int, first int) {
		first = -1
		s := Strategy{
			Kind:      "adx_trend",
			Symbol:    "AAPL",
			Fast:      5,
			Slow:      12,
			Threshold: threshold,
			StopBPS:   100,
			TargetBPS: 200,
		}
		for i, sig := range signalsOf(s, cs) {
			if sig == "buy" {
				buys++
				if first < 0 {
					first = i
				}
			}
		}
		return
	}
	low, firstLow := count(10)
	high, _ := count(50)
	if low == 0 || firstLow < 12 {
		t.Fatalf("the trend must produce a buy after the warm-up, got %d buys, first at %d", low, firstLow)
	}
	if high > low {
		t.Fatalf("a stricter ADX filter cannot add trades: %d vs %d", high, low)
	}
	if got := Signal(
		Strategy{Kind: "adx_trend", Symbol: "AAPL", Fast: 5, Slow: 12, Threshold: 10},
		cs[:8],
	); got != "hold" {
		t.Fatalf("not enough bars for the indicator must hold, got %s", got)
	}
}

func TestDCAGridKeepsInventoryWithinTheCap(t *testing.T) {
	a := armed()
	a.Policy.DrawdownBPS = 0
	a.Policy.DailyLossBPS = 500
	s := Strategy{
		Kind: "dca_grid", Symbol: "AAPL", Fast: 4, Slow: 6, Threshold: 1, StopBPS: 2000, TargetBPS: 3000,
	}
	at := time.Date(2026, 11, 4, 15, 0, 0, 0, time.UTC)
	buy := func(price string, key string) error {
		o, err := a.Reserve(s, "buy", key, D(price), at)
		if err != nil {
			return err
		}
		return a.ApplyFill(Fill{
			ID: key, OrderID: o.ID, Symbol: "AAPL", Side: "buy", Qty: o.Qty, Price: o.Limit, Time: at,
		})
	}
	if err := buy("100", "t1"); err != nil {
		t.Fatal(err)
	}
	if err := buy("99.5", "early"); err == nil {
		t.Fatal("a tranche closer than the spacing must be refused")
	}
	// The market keeps falling 1.5% per bar: tranches stop at the cap.
	price := 100.0
	for i := 2; i <= 8; i++ {
		price *= 0.985
		_ = buy(fmt.Sprintf("%.2f", price), fmt.Sprintf("t%d", i))
	}
	p := a.Positions["AAPL"]
	if p.Tranches != s.Fast {
		t.Fatalf("inventory must stop at %d tranches, got %d", s.Fast, p.Tranches)
	}
	if err := buy(fmt.Sprintf("%.2f", price*0.9), "over"); err == nil || err.Error() != "inventory cap reached" {
		t.Fatalf("the cap must refuse a further tranche, got %v", err)
	}
	if a.Exposure().GreaterThan(a.Equity().Mul(Fraction(a.Policy.ExposureBPS))) {
		t.Fatalf("exposure %s exceeds the account cap", a.Exposure())
	}
	wantStop := p.Average.Mul(D("0.8")).RoundFloor(2)
	if !p.Stop.Equal(wantStop) {
		t.Fatalf("the stop must follow the average cost: got %s want %s", p.Stop, wantStop)
	}
}

func TestEveryKindValidatesAtItsDefaultsAndBacktests(t *testing.T) {
	ds := syntheticDataset(5)
	for _, spec := range Kinds() {
		s := Strategy{ID: "k-" + spec.Kind, Name: spec.Kind, Kind: spec.Kind, Symbol: "AAPL"}.Defaults()
		if err := s.Validate(); err != nil {
			t.Fatalf("%s: defaults must be valid: %v", spec.Kind, err)
		}
		r, err := Backtest(t.Context(), ds, s, DefaultPolicy(), "all")
		if err != nil {
			t.Fatalf("%s: %v", spec.Kind, err)
		}
		if len(r.Curve) != len(ds.Candles) {
			t.Fatalf("%s: the curve must cover every bar", spec.Kind)
		}
	}
}
