package trading

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestTickRoundingFollowsTheInstrument(t *testing.T) {
	a := armed()
	s := Strategy{Kind: "ma_cross", Symbol: "EURUSD", Fast: 2, Slow: 4, Threshold: 2, StopBPS: 50, TargetBPS: 100}
	o, err := a.Reserve(s, "buy", "fx", D("1.08500"), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	// 1.085 * 0.995 = 1.079575 rounds down to the five-decimal tick; two decimals would give 1.07.
	if !o.Stop.Equal(D("1.07957")) {
		t.Fatalf("the stop must follow the 0.00001 tick, got %s", o.Stop)
	}
	if !o.Limit.Equal(D("1.08555")) {
		t.Fatalf("the limit must round up to the tick, got %s", o.Limit)
	}
	if !o.Qty.Mod(D("1000")).IsZero() {
		t.Fatalf("the quantity must be a multiple of the lot, got %s", o.Qty)
	}
}

func TestCryptoAndFXSessions(t *testing.T) {
	saturday := time.Date(2026, 11, 7, 3, 0, 0, 0, time.UTC)
	if !Session("CRYPTO", saturday) || Closing("CRYPTO", saturday.Add(20*time.Hour+50*time.Minute)) {
		t.Fatal("crypto trades around the clock and never closes")
	}
	if Session("FX", saturday) {
		t.Fatal("spot FX is closed on Saturday")
	}
	fridayLate := time.Date(2026, 11, 6, 23, 50, 0, 0, time.UTC)
	if !Closing("FX", fridayLate) {
		t.Fatal("spot FX closes out before the weekend")
	}
	if Closing("FX", time.Date(2026, 11, 5, 23, 50, 0, 0, time.UTC)) {
		t.Fatal("a Thursday night is an ordinary session for FX")
	}
	if MarketDay("CRYPTO", time.Date(2026, 11, 5, 23, 59, 0, 0, time.UTC)) != "2026-11-05" {
		t.Fatal("crypto days are UTC days")
	}
}

func cryptoCSV(bars int) string {
	var b strings.Builder
	b.WriteString("time,open,high,low,close,volume\n")
	start := time.Date(2026, 11, 6, 20, 0, 0, 0, time.UTC) // spans a weekend
	prev := 60000.0
	for i := range bars {
		next := 60000 + 900*math.Sin(float64(i)/5) + 300*math.Sin(float64(i)/1.7)
		fmt.Fprintf(&b, "%s,%.2f,%.2f,%.2f,%.2f,250\n",
			start.Add(time.Duration(i)*5*time.Minute).Format(time.RFC3339), prev,
			math.Max(prev, next)+40, math.Min(prev, next)-40, next)
		prev = next
	}
	return b.String()
}

func TestCryptoCSVImportsAcrossWeekendsAndBacktestsReproducibly(t *testing.T) {
	ds, err := ImportCSV(strings.NewReader(cryptoCSV(400)), "BTCUSDT", "fixture", "internal test")
	if err != nil {
		t.Fatal(err)
	}
	if ds.Currency != "USDT" || len(ds.Candles) != 400 || len(ds.SessionCloses) != 0 {
		t.Fatalf(
			"unexpected dataset: currency=%s bars=%d closes=%d",
			ds.Currency,
			len(ds.Candles),
			len(ds.SessionCloses),
		)
	}
	s := Strategy{
		ID: "c1", Name: "crypto", Kind: "ma_cross", Symbol: "BTCUSDT", Fast: 3, Slow: 9, Threshold: 2,
		StopBPS: 80, TargetBPS: 160,
	}
	a, err := Backtest(context.Background(), ds, s, DefaultPolicy(), "all")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := Backtest(context.Background(), ds, s, DefaultPolicy(), "all")
	if !reflect.DeepEqual(a.Curve, b.Curve) || !reflect.DeepEqual(a.Metrics, b.Metrics) {
		t.Fatal("a crypto backtest must reproduce")
	}
	if a.Metrics.Fills == 0 {
		t.Fatal("the fixture should trade")
	}
	if a.Metrics.ExposurePct <= 0 {
		t.Fatalf("a crypto position is held across bars, exposure=%v", a.Metrics.ExposurePct)
	}
}

// fxRows returns three 5-minute EURUSD candles starting at start.
func fxRows(start time.Time) string {
	var b strings.Builder
	b.WriteString("time,open,high,low,close,volume\n")
	for i := range 3 {
		fmt.Fprintf(
			&b,
			"%s,1.08500,1.08600,1.08400,1.08550,100000\n",
			start.Add(time.Duration(i)*5*time.Minute).Format(time.RFC3339),
		)
	}
	return b.String()
}

func TestFXCSVImportRespectsTheWeekend(t *testing.T) {
	fridayEvening := time.Date(2026, 11, 6, 23, 55, 0, 0, time.UTC)
	rows := fxRows(fridayEvening)
	if _, err := ImportCSV(strings.NewReader(rows), "EURUSD", "fixture", "internal test"); err == nil {
		t.Fatal("candles on a Saturday must be rejected for spot FX")
	}
	weekday := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	ok := fxRows(weekday)
	if _, err := ImportCSV(strings.NewReader(ok), "EURUSD", "fixture", "internal test"); err != nil {
		t.Fatalf("weekday FX candles must import: %v", err)
	}
	if _, err := ImportCSV(strings.NewReader(ok), "AAPL", "fixture", "internal test"); err == nil {
		t.Fatal("US equities use the Alpaca feed, not CSV")
	}
}

func TestCoverageMatrixNamesEveryMarket(t *testing.T) {
	seen := map[string]Coverage{}
	for _, c := range CoverageMatrix() {
		if c.Data == "" || c.Execution == "" || c.Source == "" || c.Notes == "" {
			t.Fatalf("incomplete coverage row: %+v", c)
		}
		seen[c.Market] = c
	}
	for _, i := range Instruments {
		if _, ok := seen[i.Market]; !ok {
			t.Fatalf("market %s of %s has no coverage row", i.Market, i.Symbol)
		}
	}
	if seen["CRYPTO"].Execution != "backtest_only" || seen["FX"].Execution != "backtest_only" {
		t.Fatal("crypto and FX must say they only backtest")
	}
}
