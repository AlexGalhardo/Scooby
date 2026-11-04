package trading

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"slices"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "rewrite golden files")

// syntheticDataset builds several US sessions of 5-minute bars that swing enough to produce trades.
func syntheticDataset(sessions int) Dataset {
	var cs []Candle
	start := time.Date(2026, 10, 26, 14, 0, 0, 0, time.UTC)
	prev := 100.0
	for day := range sessions {
		for j := range 70 {
			at := start.AddDate(0, 0, day).Add(time.Duration(j) * 5 * time.Minute)
			n := float64(day*70 + j)
			closeP := 100 + 4*math.Sin(n/7) + 2*math.Sin(n/3.1) + 0.01*n
			closeP = math.Round(closeP*100) / 100
			hi := math.Max(prev, closeP) + 0.6
			lo := math.Min(prev, closeP) - 0.6
			cs = append(cs, Candle{
				Time: at, Open: D(fmt.Sprintf("%.2f", prev)), High: D(fmt.Sprintf("%.2f", hi)),
				Low: D(fmt.Sprintf("%.2f", lo)), Close: D(fmt.Sprintf("%.2f", closeP)), Volume: D("1000000"),
			})
			prev = closeP
		}
	}
	return Dataset{Symbol: "AAPL", Hash: DatasetHash(cs), Candles: cs}
}

func fastStrategy() Strategy {
	return Strategy{
		ID: "s1", Name: "stats", Kind: "ma_cross", Symbol: "AAPL", Fast: 3, Slow: 9,
		Threshold: 2, StopBPS: 60, TargetBPS: 120,
	}
}

func TestStatsAreComputedAndStated(t *testing.T) {
	r, err := Backtest(context.Background(), syntheticDataset(5), fastStrategy(), DefaultPolicy(), "all")
	if err != nil {
		t.Fatal(err)
	}
	m := r.Metrics
	for name, v := range map[string]float64{
		"sharpe": m.Sharpe, "sortino": m.Sortino, "calmar": m.Calmar, "sqn": m.SQN, "annual": m.AnnualReturnPct,
	} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatalf("%s is not finite: %v", name, v)
		}
	}
	if m.PeriodsPerYear <= 0 || m.RiskFreeAnnualPct != RiskFreeAnnualPct {
		t.Fatalf("the report must state its assumptions: periods=%v rf=%v", m.PeriodsPerYear, m.RiskFreeAnnualPct)
	}
	if m.ExposurePct < 0 || m.ExposurePct > 100 {
		t.Fatalf("exposure out of range: %v", m.ExposurePct)
	}
	if len(r.TradePnL) != m.Trades || m.Trades == 0 {
		t.Fatalf("trade P&L (%d) must match trades (%d) and the fixture must trade", len(r.TradePnL), m.Trades)
	}
}

func TestMonteCarloIsDeterministicAndLowTradesAreFlagged(t *testing.T) {
	ds := syntheticDataset(5)
	a, err := Backtest(context.Background(), ds, fastStrategy(), DefaultPolicy(), "all")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := Backtest(context.Background(), ds, fastStrategy(), DefaultPolicy(), "all")
	if a.Robustness == nil || a.Robustness.MonteCarlo == nil {
		t.Fatal("robustness with a Monte Carlo block is required")
	}
	if *a.Robustness.MonteCarlo != *b.Robustness.MonteCarlo {
		t.Fatal("the bootstrap must be reproducible")
	}
	mc := a.Robustness.MonteCarlo
	if !(mc.P5Pct <= mc.P50Pct && mc.P50Pct <= mc.P95Pct) || mc.ProbLoss < 0 || mc.ProbLoss > 1 {
		t.Fatalf("inconsistent percentiles: %+v", mc)
	}
	if a.Metrics.Trades < MinTrades && !a.Robustness.LowTrades {
		t.Fatal("a run with fewer than 30 trades must be flagged")
	}
}

func TestCostStressNeverImprovesTheResult(t *testing.T) {
	r, err := Backtest(context.Background(), syntheticDataset(5), fastStrategy(), DefaultPolicy(), "all")
	if err != nil {
		t.Fatal(err)
	}
	if r.Robustness.StressReturn == nil {
		t.Fatalf("cost stress must run: %s", r.Robustness.StressNote)
	}
	if r.Robustness.StressReturn.GreaterThan(r.Metrics.ReturnPct) {
		t.Fatalf("doubled costs improved the result: %s > %s", r.Robustness.StressReturn, r.Metrics.ReturnPct)
	}
}

func TestDeflatedSharpeDiscountsTheNumberOfTrials(t *testing.T) {
	one := DeflatedSharpe(0.05, 500, 0, 3, 1)
	many := DeflatedSharpe(0.05, 500, 0, 3, 200)
	if !(many < one) || one <= 0 || one >= 1 {
		t.Fatalf("more trials must lower the probability: one=%v many=%v", one, many)
	}
	if DeflatedSharpe(0.05, 2, 0, 3, 1) != 0 {
		t.Fatal("too few returns must give zero, not a guess")
	}
}

func TestPartitionsAreTrainTestValidation(t *testing.T) {
	ds := syntheticDataset(5)
	n := len(ds.Candles)
	want := map[string]int{"train": n * 6 / 10, "test": n*8/10 - n*6/10, "validation": n - n*8/10, "all": n}
	for partition, size := range want {
		r, err := Backtest(context.Background(), ds, fastStrategy(), DefaultPolicy(), partition)
		if err != nil {
			t.Fatalf("%s: %v", partition, err)
		}
		if len(r.Curve) != size {
			t.Fatalf("%s: %d bars, want %d", partition, len(r.Curve), size)
		}
	}
	if _, err := Backtest(context.Background(), ds, fastStrategy(), DefaultPolicy(), "holdout"); err == nil {
		t.Fatal("an unknown partition must be rejected")
	}
}

func TestWalkForwardReportsEveryWindow(t *testing.T) {
	ds := syntheticDataset(8)
	r, err := Backtest(context.Background(), ds, fastStrategy(), DefaultPolicy(), "walkforward")
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Windows) != WalkForwardWindows {
		t.Fatalf("want %d windows, got %d", WalkForwardWindows, len(r.Windows))
	}
	bars := 0
	for _, w := range r.Windows {
		bars += w.Bars
		if w.Bars == 0 {
			t.Fatalf("window %d is empty", w.Index)
		}
	}
	if bars != len(ds.Candles) {
		t.Fatalf("windows must cover the whole sample: %d of %d bars", bars, len(ds.Candles))
	}
	if r.Partition != "walkforward" {
		t.Fatalf("the report must name its partition, got %q", r.Partition)
	}
}

// perturbAfter rescales every candle after index k, which a causal engine must never notice before k.
func perturbAfter(ds Dataset, k int) Dataset {
	out := ds
	out.Candles = slices.Clone(ds.Candles)
	for i := k + 1; i < len(out.Candles); i++ {
		f := D(fmt.Sprintf("%.4f", 1+0.15*math.Sin(float64(i))))
		c := out.Candles[i]
		c.Open, c.High, c.Low, c.Close = c.Open.Mul(f), c.High.Mul(f), c.Low.Mul(f), c.Close.Mul(f)
		out.Candles[i] = c
	}
	out.Hash = DatasetHash(out.Candles)
	return out
}

func TestNoStrategyLooksAhead(t *testing.T) {
	ds := syntheticDataset(5)
	k := len(ds.Candles) / 2
	future := perturbAfter(ds, k)
	for _, spec := range Kinds() {
		s := fastStrategy()
		s.Kind = spec.Kind
		for i := s.Slow; i <= k; i++ {
			if Signal(s, ds.Candles[:i+1]) != Signal(s, future.Candles[:i+1]) {
				t.Fatalf("%s: the signal at bar %d changed with the future", spec.Kind, i)
			}
		}
		a, err := runBacktest(context.Background(), ds, s, DefaultPolicy(), "all")
		if err != nil {
			t.Fatal(err)
		}
		b, err := runBacktest(context.Background(), future, s, DefaultPolicy(), "all")
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i <= k; i++ {
			if !a.Curve[i].Equity.Equal(b.Curve[i].Equity) {
				t.Fatalf("%s: equity at bar %d depends on later candles", spec.Kind, i)
			}
		}
		before := func(r Result) []Fill {
			var out []Fill
			for _, f := range r.Fills {
				if !f.Time.After(ds.Candles[k].Time) {
					out = append(out, f)
				}
			}
			return out
		}
		if len(before(a)) != len(before(b)) {
			t.Fatalf("%s: fills before bar %d changed with the future", spec.Kind, k)
		}
	}
}

func TestBacktestMetricsGolden(t *testing.T) {
	r, err := Backtest(context.Background(), syntheticDataset(5), fastStrategy(), DefaultPolicy(), "all")
	if err != nil {
		t.Fatal(err)
	}
	round := func(v float64) float64 { return math.Round(v*1e6) / 1e6 }
	got, _ := json.MarshalIndent(map[string]any{
		"return_pct": r.Metrics.ReturnPct.StringFixed(4), "trades": r.Metrics.Trades, "fills": r.Metrics.Fills,
		"max_drawdown_pct": r.Metrics.MaxDrawdownPct.StringFixed(4), "costs": r.Metrics.Costs.StringFixed(2),
		"sharpe": round(r.Metrics.Sharpe), "sortino": round(r.Metrics.Sortino), "calmar": round(r.Metrics.Calmar),
		"sqn": round(r.Metrics.SQN), "exposure_pct": round(r.Metrics.ExposurePct),
		"benchmark_return_pct": r.Metrics.BenchmarkReturnPct.StringFixed(4),
	}, "", "  ")
	got = append(got, '\n')
	const path = "testdata/backtest_metrics.golden.json"
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing golden file (run with -update): %v", err)
	}
	if string(want) != string(got) {
		t.Fatalf("metrics changed.\nwant:\n%s\ngot:\n%s", want, got)
	}
}
