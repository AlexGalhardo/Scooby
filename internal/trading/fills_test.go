package trading

import (
	"testing"
	"time"
)

func bar(at time.Time, open, high, low, vol string) Candle {
	return Candle{Time: at, Open: D(open), High: D(high), Low: D(low), Close: D(open), Volume: D(vol)}
}

func buyOrder(t *testing.T, a *State, qty, limit string) *Order {
	t.Helper()
	o, err := a.Reserve(strategy(), "buy", "k", D("100"), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	o.Qty, o.Limit = D(qty), D(limit)
	o.Reserved = o.Qty.Mul(o.Limit)
	return o
}

func TestRestingBuyFillsOnlyWhenTheBarTradesThrough(t *testing.T) {
	a := armed()
	o := buyOrder(t, &a, "10", "102")
	at := time.Date(2026, 11, 4, 15, 0, 0, 0, time.UTC)

	// The open is above the limit and the low only touches it: no fill, the order keeps resting.
	if err := Simulate(&a, o, bar(at, "105", "106", "102", "100000")); err != nil {
		t.Fatal(err)
	}
	if !o.Filled.IsZero() || !o.Open() {
		t.Fatalf("a touch must not fill and must not cancel: filled=%s status=%s", o.Filled, o.Status)
	}

	// The next bar trades strictly through the limit: it fills at the limit, not at the better price.
	if err := Simulate(&a, o, bar(at.Add(5*time.Minute), "105", "106", "101", "100000")); err != nil {
		t.Fatal(err)
	}
	if o.Status != "filled" || !o.Average.Equal(D("102")) {
		t.Fatalf("expected a fill at the limit, got status=%s average=%s", o.Status, o.Average)
	}
}

func TestMarketableBuyNeverPaysAboveTheBarHigh(t *testing.T) {
	a := armed()
	o := buyOrder(t, &a, "10", "101")
	at := time.Date(2026, 11, 4, 15, 0, 0, 0, time.UTC)
	// Open equals the high, so open plus slippage would print above the bar.
	if err := Simulate(&a, o, bar(at, "100", "100", "99", "100000")); err != nil {
		t.Fatal(err)
	}
	if o.Status != "filled" || o.Average.GreaterThan(D("100")) {
		t.Fatalf("the price must be capped at the high: status=%s average=%s", o.Status, o.Average)
	}
}

func TestPartialFillSpansSeveralBars(t *testing.T) {
	a := armed()
	o := buyOrder(t, &a, "100", "101")
	at := time.Date(2026, 11, 4, 15, 0, 0, 0, time.UTC)
	// 1% of 4000 shares is 40 per bar: 40, 40, then the last 20.
	want := []struct {
		status string
		filled string
	}{{"partially_filled", "40"}, {"partially_filled", "80"}, {"filled", "100"}}
	for i, w := range want {
		if err := Simulate(&a, o, bar(at.Add(time.Duration(i)*5*time.Minute), "100", "101", "99", "4000")); err != nil {
			t.Fatal(err)
		}
		if o.Status != w.status || !o.Filled.Equal(D(w.filled)) {
			t.Fatalf("bar %d: status=%s filled=%s, want %s %s", i, o.Status, o.Filled, w.status, w.filled)
		}
	}
	if len(a.Fills) != 3 {
		t.Fatalf("expected three fills, got %d", len(a.Fills))
	}
	if !a.Positions["AAPL"].Qty.Equal(D("100")) {
		t.Fatalf("the position must hold the filled quantity, got %s", a.Positions["AAPL"].Qty)
	}
}

func TestThinBarLeavesTheOrderOpen(t *testing.T) {
	a := armed()
	o := buyOrder(t, &a, "10", "101")
	// 1% of 50 shares rounds down to zero: nothing fills, nothing is canceled.
	if err := Simulate(&a, o, bar(time.Now(), "100", "101", "99", "50")); err != nil {
		t.Fatal(err)
	}
	if !o.Open() || !o.Filled.IsZero() {
		t.Fatalf("a thin bar must leave the order open, status=%s", o.Status)
	}
}
