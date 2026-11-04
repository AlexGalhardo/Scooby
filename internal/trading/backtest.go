package trading

import (
	"context"
	"errors"
	"fmt"
	"time"

	d "github.com/shopspring/decimal"
)

type Point struct {
	Time      time.Time `json:"time"`
	Equity    d.Decimal `json:"equity"`
	Benchmark d.Decimal `json:"benchmark"`
}
type Metrics struct {
	BenchmarkReturnPct d.Decimal `json:"benchmark_return_pct"`
	TurnoverPct        d.Decimal `json:"turnover_pct"`
	AverageWin         d.Decimal `json:"average_win"`
	AverageLoss        d.Decimal `json:"average_loss"`
	MaxDrawdownSeconds int64     `json:"max_drawdown_seconds"`

	Sharpe            float64 `json:"sharpe"`
	Sortino           float64 `json:"sortino"`
	Calmar            float64 `json:"calmar"`
	SQN               float64 `json:"sqn"`
	AnnualReturnPct   float64 `json:"annual_return_pct"`
	ExposurePct       float64 `json:"exposure_pct"`
	RiskFreeAnnualPct float64 `json:"risk_free_annual_pct"`
	PeriodsPerYear    float64 `json:"periods_per_year"`

	ReturnPct      d.Decimal `json:"return_pct"`
	MaxDrawdownPct d.Decimal `json:"max_drawdown_pct"`
	Realized       d.Decimal `json:"realized"`
	Costs          d.Decimal `json:"costs"`
	Fills          int       `json:"fills"`
	Trades         int       `json:"trades"`
	WinRate        float64   `json:"win_rate"`
	ExposureBars   int       `json:"exposure_bars"`
}
type Result struct {
	CalendarSource string      `json:"calendar_source"`
	Metrics        Metrics     `json:"metrics"`
	Curve          []Point     `json:"curve"`
	Fills          []Fill      `json:"fills"`
	DatasetHash    string      `json:"dataset_hash"`
	Strategy       Strategy    `json:"strategy"`
	Policy         Policy      `json:"policy"`
	Warnings       []string    `json:"warnings"`
	Final          State       `json:"-"`
	EngineVersion  string      `json:"engine_version"`
	Partition      string      `json:"partition"`
	TradePnL       []d.Decimal `json:"trade_pnl"`
	Windows        []Window    `json:"windows,omitempty"`
	Robustness     *Robustness `json:"robustness,omitempty"`
}

// Backtest replays the strategy on a partition of the dataset and reports the result with its
// robustness checks. Partitions: train (first 60%), test (next 20%), validation (last 20%), all, and
// walkforward (all, plus consecutive windows).
func Backtest(ctx context.Context, ds Dataset, s Strategy, policy Policy, partition string) (Result, error) {
	return BacktestTrials(ctx, ds, s, policy, partition, 1)
}

// BacktestTrials is Backtest for a configuration that is one of trials tried; the deflated Sharpe
// ratio discounts the result for that.
func BacktestTrials(
	ctx context.Context, ds Dataset, s Strategy, policy Policy, partition string, trials int,
) (Result, error) {
	run := partition
	if partition == "walkforward" {
		run = "all"
	}
	r, sr, n, skew, kurt, err := runBacktestStats(ctx, ds, s, policy, run)
	if err != nil {
		return r, err
	}
	r.Partition = partition
	r.Robustness = robustness(ctx, ds, s, policy, run, r, trials, sr, n, skew, kurt)
	if r.Robustness.LowTrades {
		r.Warnings = append(r.Warnings, "Fewer than 30 trades: the statistics are weak.")
	}
	if partition == "walkforward" {
		r.Windows = walkForward(ctx, ds, s, policy)
		r.Warnings = append(r.Warnings, "Walk-forward windows start cold, each with its own warm-up.")
	}
	return r, nil
}

func runBacktest(ctx context.Context, ds Dataset, s Strategy, policy Policy, partition string) (Result, error) {
	r, _, _, _, _, err := runBacktestStats(ctx, ds, s, policy, partition)
	return r, err
}

func walkForward(ctx context.Context, ds Dataset, s Strategy, policy Policy) []Window {
	size := len(ds.Candles) / WalkForwardWindows
	out := make([]Window, 0, WalkForwardWindows)
	for w := range WalkForwardWindows {
		end := (w + 1) * size
		if w == WalkForwardWindows-1 {
			end = len(ds.Candles)
		}
		sub := ds
		sub.Candles = ds.Candles[w*size : end]
		win := Window{Index: w + 1, Bars: len(sub.Candles)}
		res, err := runBacktest(ctx, sub, s, policy, "all")
		if err != nil {
			win.Note = err.Error()
			out = append(out, win)
			continue
		}
		win.Trades = res.Metrics.Trades
		win.ReturnPct = res.Metrics.ReturnPct
		win.MaxDrawdownPct = res.Metrics.MaxDrawdownPct
		win.Enough = win.Trades >= MinTradesPerWindow
		out = append(out, win)
	}
	return out
}

func runBacktestStats(
	ctx context.Context, ds Dataset, s Strategy, policy Policy, partition string,
) (Result, float64, int, float64, float64, error) {
	r, err := backtestCore(ctx, ds, s, policy, partition)
	if err != nil {
		return r, 0, 0, 0, 0, err
	}
	inst, _ := FindInstrument(s.Symbol)
	sr, n, skew, kurt := applyStats(&r, inst.Market, Initial())
	return r, sr, n, skew, kurt, nil
}

func backtestCore(ctx context.Context, ds Dataset, s Strategy, policy Policy, partition string) (Result, error) {
	r := Result{
		CalendarSource: ds.CalendarSource,
		Curve:          []Point{},
		Warnings: []string{
			"Simulation: OHLC does not reveal intrabar sequence; stop takes priority over target.",
			"Costs in bps; taxes not included. No profitability guarantee.",
			"Gross buy-and-hold benchmark; remaining positions are closed at the end of the sample.",
		},
		DatasetHash:   ds.Hash,
		Strategy:      s,
		Policy:        policy,
		EngineVersion: BuildVersion,
		Partition:     partition,
	}
	if err := s.Validate(); err != nil {
		return r, err
	}
	if err := policy.Validate(); err != nil {
		return r, err
	}
	if ds.Symbol != s.Symbol {
		return r, errors.New("dataset does not match the symbol")
	}
	cs := ds.Candles
	trainEnd, testEnd := len(cs)*6/10, len(cs)*8/10
	switch partition {
	case "train":
		cs = cs[:trainEnd]
	case "test":
		cs = cs[trainEnd:testEnd]
	case "validation":
		cs = cs[testEnd:]
	case "all":
		r.Warnings = append(r.Warnings, "Exploratory: uses the full period, not out-of-sample testing.")
	default:
		return r, errors.New("partition must be train, test, validation or all")
	}
	if len(cs) < s.Slow+2 {
		return r, errors.New("insufficient history for warmup and evaluation")
	}
	i, _ := FindInstrument(s.Symbol)
	a := NewState("backtest", i.Currency, "backtest")
	a.Halted = false
	a.Policy = policy
	pending := ""
	peak := a.Initial
	wins, trades := 0, 0
	cycleOpen := false
	cycleStart := Zero
	winTotal, lossTotal := Zero, Zero
	peakAt := cs[0].Time
	for idx, c := range cs {
		if err := ctx.Err(); err != nil {
			return r, err
		}
		if err := c.Validate(); err != nil {
			return r, err
		}
		day := MarketDay(i.Market, c.Time)
		if a.Day != day {
			a.Day = day
			a.DayStart = a.Equity()
		}
		if pending != "" {
			// A resting or partly filled order stays open for the rest of its session (ADR 0016).
			o := a.Order(pending)
			switch {
			case o == nil || !o.Open():
				pending = ""
			case MarketDay(i.Market, o.CreatedAt) != day:
				o.Status = "expired"
				o.Reserved = Zero
				pending = ""
			default:
				if err := Simulate(&a, o, c); err != nil {
					return r, err
				}
				if !o.Open() {
					pending = ""
				}
			}
		}
		if !cycleOpen && a.Positions[s.Symbol].Qty.IsPositive() {
			cycleOpen = true
			cycleStart = a.Realized
		}
		a.Mark(s.Symbol, c.Close)
		if p, ok := a.Positions[s.Symbol]; ok {
			r.Metrics.ExposureBars++
			exit := Zero
			// Gap-through stops execute at the worse opening price. Never choose target first.
			if c.Low.LessThanOrEqual(p.Stop) {
				exit = d.Min(c.Open, p.Stop)
			} else if c.High.GreaterThanOrEqual(p.Target) {
				exit = p.Target
			} else if ds.ClosingAt(c.Time) || idx == len(cs)-1 {
				exit = c.Close
			}
			if exit.IsPositive() {
				o := Order{
					ID:        ID(),
					Key:       fmt.Sprint(idx, "/exit"),
					Symbol:    s.Symbol,
					Side:      "sell",
					Qty:       p.Qty,
					Status:    "approved",
					CreatedAt: c.Time,
				}
				a.Orders = append(a.Orders, o)
				price := exit.Mul(D("1").Sub(Fraction(policy.SlippageBPS)))
				if err := a.ApplyFill(
					Fill{
						ID:      o.ID,
						OrderID: o.ID,
						Symbol:  s.Symbol,
						Side:    "sell",
						Qty:     p.Qty,
						Price:   price,
						Fees:    p.Qty.Mul(price).Mul(Fraction(policy.CostBPS)),
						Time:    c.Time,
					},
				); err != nil {
					return r, err
				}
			}
		}
		if cycleOpen && !a.Positions[s.Symbol].Qty.IsPositive() {
			gain := a.Realized.Sub(cycleStart)
			r.TradePnL = append(r.TradePnL, gain)
			if gain.IsPositive() {
				wins++
				winTotal = winTotal.Add(gain)
			} else {
				lossTotal = lossTotal.Add(gain)
			}
			trades++
			cycleOpen = false
		}
		eq := a.Equity()
		if eq.GreaterThanOrEqual(peak) {
			peakAt = c.Time
		}
		peak = d.Max(peak, eq)
		if eq.LessThan(peak) {
			seconds := int64(c.Time.Sub(peakAt).Seconds())
			if seconds > r.Metrics.MaxDrawdownSeconds {
				r.Metrics.MaxDrawdownSeconds = seconds
			}
		}
		dd := peak.Sub(eq).Div(peak).Mul(D("100"))
		r.Metrics.MaxDrawdownPct = d.Max(r.Metrics.MaxDrawdownPct, dd)
		r.Curve = append(r.Curve, Point{c.Time, eq, a.Initial.Mul(c.Close).Div(cs[0].Open)})
		if idx >= s.Slow && idx < len(cs)-1 && !ds.ClosingAt(c.Time) && Session(i.Market, c.Time) {
			side := Signal(s, cs[:idx+1])
			if side != "hold" {
				o, err := a.Reserve(s, side, fmt.Sprint(idx, "/signal"), c.Close, c.Time)
				if err == nil {
					pending = o.ID
				}
			}
		}
	}
	r.Metrics.ReturnPct = a.Equity().Sub(a.Initial).Div(a.Initial).Mul(D("100"))
	r.Metrics.Realized = a.Realized
	r.Metrics.Fills = len(a.Fills)
	r.Metrics.Trades = trades
	if trades > 0 {
		r.Metrics.WinRate = float64(wins) / float64(trades) * 100
	}
	for _, f := range a.Fills {
		r.Metrics.Costs = r.Metrics.Costs.Add(f.Fees)
		r.Metrics.TurnoverPct = r.Metrics.TurnoverPct.Add(f.Qty.Mul(f.Price).Div(a.Initial).Mul(D("100")))
	}
	if wins > 0 {
		r.Metrics.AverageWin = winTotal.Div(d.NewFromInt(int64(wins)))
	}
	if trades > wins {
		r.Metrics.AverageLoss = lossTotal.Div(d.NewFromInt(int64(trades - wins)))
	}
	r.Metrics.BenchmarkReturnPct = cs[len(cs)-1].Close.Sub(cs[0].Open).Div(cs[0].Open).Mul(D("100"))
	r.Fills = a.Fills
	r.Final = a
	r.Warnings = append(r.Warnings, DatasetProblems(ds)...)
	return r, nil
}
