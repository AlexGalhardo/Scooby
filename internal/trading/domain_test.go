package trading

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
)

func strategy() Strategy {
	return Strategy{Kind: "ma_cross", Symbol: "AAPL", Fast: 2, Slow: 4, Threshold: 2, StopBPS: 100, TargetBPS: 200}
}
func TestRiskReservationAndIdempotentFill(t *testing.T) {
	a := NewState("test", "USD", "simulation")
	a.Halted = false
	s := strategy()
	at := time.Now()
	o, err := a.Reserve(s, "buy", "signal-1", D("100"), at)
	if err != nil {
		t.Fatal(err)
	}
	id := o.ID
	qty := o.Qty
	if a.Exposure().GreaterThan(a.Initial.Mul(Fraction(a.Policy.ExposureBPS))) {
		t.Fatal("exposure exceeded")
	}
	again, err := a.Reserve(s, "buy", "signal-1", D("100"), at)
	if err != nil || again.ID != id || len(a.Orders) != 1 {
		t.Fatal("duplicate intent")
	}
	half := qty.Div(D("2")).Floor()
	f := Fill{ID: "partial-1", OrderID: id, Symbol: "AAPL", Side: "buy", Qty: half, Price: D("100"), Fees: D("1"), Time: at}
	if err := a.ApplyFill(f); err != nil {
		t.Fatal(err)
	}
	cash := a.Cash
	if err := a.ApplyFill(f); err != nil || !a.Cash.Equal(cash) {
		t.Fatal("duplicate fill changed cash")
	}
	if a.Order(id).Status != "partially_filled" || !a.Order(id).Reserved.IsPositive() {
		t.Fatal("partial did not preserve reservation")
	}
	f.ID = "overfill"
	f.Qty = qty
	if err := a.ApplyFill(f); err == nil {
		t.Fatal("overfill accepted")
	}
	a.Halt("test")
	s.Symbol = "NVDA"
	if _, err := a.Reserve(s, "buy", "new", D("100"), at); err == nil {
		t.Fatal("halt ignored")
	}
	s.Symbol = "AAPL"
	a.Order(id).Status = "canceled"
	a.Order(id).Reserved = Zero
	if _, err := a.Reserve(s, "sell", "exit", D("100"), at); err != nil {
		t.Fatal("risk reduction blocked", err)
	}
}
func TestDailyLossPersistsAndDecimals(t *testing.T) {
	a := NewState("test", "USD", "simulation")
	a.Halted = false
	a.Cash = D("98999.99")
	if _, err := a.Reserve(strategy(), "buy", "loss", D("100"), time.Now()); err == nil || !a.Halted {
		t.Fatal("daily loss not enforced")
	}
	if !D("0.1").Add(D("0.2")).Equal(D("0.3")) {
		t.Fatal("inexact decimal")
	}
}
func TestCSVRejectsAmbiguity(t *testing.T) {
	sample := SampleCSV()
	for _, invalid := range []string{strings.Replace(sample, "30.00", "NaN", 1), strings.Replace(sample, "30.00", "1e1000000000", 1), strings.Replace(sample, "10:05:00-03:00", "10:00:00-03:00", 1), strings.Replace(sample, "10:05:00-03:00", "10:05:00", 1), strings.Replace(sample, "30.20,29.90", "29.00,31.00", 1)} {
		if _, err := ImportCSV(strings.NewReader(invalid), "PETR4", "test", "test license"); err == nil {
			t.Fatal("invalid CSV accepted")
		}
	}
	ds, err := ImportCSV(strings.NewReader(sample), "PETR4", "test", "test license")
	if err != nil || len(ds.Candles) != 3 || len(ds.Hash) != 64 {
		t.Fatal(ds, err)
	}
}
func TestSignalsAndBacktestReproducibility(t *testing.T) {
	s := strategy()
	cs := []Candle{}
	start := time.Date(2026, 10, 26, 14, 0, 0, 0, time.UTC)
	for j := 0; j < 70; j++ {
		price := D("100").Add(D([]string{"0", "1", "2", "1", "0", "-1", "-2", "-1"}[j%8]))
		open := price
		if j > 0 {
			open = cs[j-1].Close
		}
		cs = append(cs, Candle{start.Add(time.Duration(j) * 5 * time.Minute), open, price.Add(D("3")), price.Sub(D("3")), price, D("1000000")})
	}
	ds := Dataset{Symbol: "AAPL", Hash: DatasetHash(cs), Candles: cs}
	r, err := Backtest(context.Background(), ds, s, DefaultPolicy(), "all")
	if err != nil {
		t.Fatal(err)
	}
	r2, err := Backtest(context.Background(), ds, s, DefaultPolicy(), "all")
	if err != nil || !reflect.DeepEqual(r.Curve, r2.Curve) || !reflect.DeepEqual(r.Metrics, r2.Metrics) {
		t.Fatal("not reproducible", err)
	}
	if r.Metrics.Fills == 0 {
		t.Fatal("fixture should produce fills")
	}
	for _, f := range r.Fills {
		if f.Time.Equal(cs[0].Time) {
			t.Fatal("warmup/lookahead violated")
		}
	}
	s.Kind = "mean_reversion"
	flat := append([]Candle(nil), cs...)
	for i := range flat {
		flat[i].Close = D("100")
	}
	if Signal(s, flat) != "hold" {
		t.Fatal("zero variance must not trade")
	}
}
func TestMarketTimezoneDST(t *testing.T) {
	if !Session("US", time.Date(2026, 11, 2, 14, 30, 0, 0, time.UTC)) {
		t.Fatal("US standard time opening")
	}
	if Session("US", time.Date(2026, 11, 2, 13, 30, 0, 0, time.UTC)) {
		t.Fatal("DST incorrect")
	}
	if Session("B3", time.Date(2026, 11, 1, 14, 0, 0, 0, time.UTC)) {
		t.Fatal("Sunday session")
	}
}

func TestB3CalendarHolidaysAndScheduleChange(t *testing.T) {
	parse := func(s string) time.Time {
		v, e := time.Parse(time.RFC3339, s)
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	if Session("B3", parse("2026-11-02T14:00:00-03:00")) {
		t.Fatal("Finados must be closed")
	}
	if Session("B3", parse("2026-02-18T10:00:00-03:00")) {
		t.Fatal("Ash Wednesday opens at 13h")
	}
	if !Session("B3", parse("2026-11-03T17:30:00-03:00")) {
		t.Fatal("November schedule not applied")
	}
	if Session("B3", parse("2026-10-26T17:30:00-03:00")) {
		t.Fatal("old schedule not applied")
	}
	if Session("B3", parse("2027-01-04T14:00:00-03:00")) {
		t.Fatal("unreviewed year must fail closed")
	}
}
