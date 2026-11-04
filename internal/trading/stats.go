package trading

import (
	"context"
	"hash/fnv"
	"math"
	"slices"

	d "github.com/shopspring/decimal"
)

// Statistical reporting settings. Money stays decimal; these statistics are floating point because
// they are estimates, not ledger values.
const (
	// RiskFreeAnnualPct is the annual risk-free rate used by Sharpe and Sortino. Zero keeps them
	// comparable across currencies; the report states it.
	RiskFreeAnnualPct = 0.0
	tradingDays       = 252
	bootstrapRuns     = 1000
	bootstrapBlock    = 3
	// MinTrades is the number of trades below which a run is flagged as statistically weak.
	MinTrades = 30
	// WalkForwardWindows is the number of consecutive windows of the walk-forward report.
	WalkForwardWindows = 4
	// MinTradesPerWindow is the number of trades below which a window is flagged as too thin.
	MinTradesPerWindow = 3
)

// Window summarises one consecutive slice of a walk-forward run. Each window starts cold, with its
// own warm-up, so it measures how the rules behave on data they were not tuned on.
type Window struct {
	Index          int       `json:"index"`
	Bars           int       `json:"bars"`
	Trades         int       `json:"trades"`
	ReturnPct      d.Decimal `json:"return_pct"`
	MaxDrawdownPct d.Decimal `json:"max_drawdown_pct"`
	Enough         bool      `json:"enough_trades"`
	Note           string    `json:"note,omitempty"`
}

// MonteCarlo is a block bootstrap of the closed trades: it shows how much of the result depends on
// the order the trades happened in.
type MonteCarlo struct {
	Runs      int     `json:"runs"`
	Block     int     `json:"block"`
	P5Pct     float64 `json:"p5_return_pct"`
	P50Pct    float64 `json:"p50_return_pct"`
	P95Pct    float64 `json:"p95_return_pct"`
	ProbLoss  float64 `json:"probability_of_loss"`
	SampleOf  int     `json:"trades"`
	Seed      uint64  `json:"seed"`
	Technique string  `json:"technique"`
}

// Robustness gathers the checks that guard against overfitting.
type Robustness struct {
	Trials         int         `json:"trials"`
	LowTrades      bool        `json:"low_trades"`
	DeflatedSharpe float64     `json:"deflated_sharpe"`
	MonteCarlo     *MonteCarlo `json:"monte_carlo,omitempty"`
	StressReturn   *d.Decimal  `json:"cost_stress_return_pct,omitempty"`
	StressNote     string      `json:"cost_stress_note,omitempty"`
}

// returnsOf converts an equity curve into per-bar simple returns.
func returnsOf(curve []Point) []float64 {
	out := make([]float64, 0, len(curve))
	for i := 1; i < len(curve); i++ {
		prev := curve[i-1].Equity
		if !prev.IsPositive() {
			continue
		}
		out = append(out, curve[i].Equity.Div(prev).Sub(D("1")).InexactFloat64())
	}
	return out
}

func moments(xs []float64) (mean, sd, skew, kurt float64) {
	n := float64(len(xs))
	if n < 2 {
		return
	}
	for _, x := range xs {
		mean += x
	}
	mean /= n
	var m2, m3, m4 float64
	for _, x := range xs {
		v := x - mean
		m2 += v * v
		m3 += v * v * v
		m4 += v * v * v * v
	}
	m2, m3, m4 = m2/n, m3/n, m4/n
	sd = math.Sqrt(m2 * n / (n - 1))
	if m2 > 0 {
		skew = m3 / math.Pow(m2, 1.5)
		kurt = m4 / (m2 * m2)
	}
	return
}

// periodsPerYear estimates how many bars a year holds from the sample itself.
func periodsPerYear(curve []Point, market string) float64 {
	days := map[string]struct{}{}
	for _, p := range curve {
		days[MarketDay(market, p.Time)] = struct{}{}
	}
	if len(days) == 0 {
		return 0
	}
	return float64(len(curve)) / float64(len(days)) * tradingDays
}

// applyStats fills the ratio metrics of a finished run and returns the per-bar Sharpe ratio, the
// number of returns and the return skew and kurtosis that the deflated Sharpe needs.
func applyStats(r *Result, market string, initial d.Decimal) (sr float64, n int, skew, kurt float64) {
	rets := returnsOf(r.Curve)
	n = len(rets)
	ppy := periodsPerYear(r.Curve, market)
	r.Metrics.RiskFreeAnnualPct = RiskFreeAnnualPct
	r.Metrics.PeriodsPerYear = ppy
	if n > 0 && len(r.Curve) > 0 {
		r.Metrics.ExposurePct = float64(r.Metrics.ExposureBars) / float64(len(r.Curve)) * 100
	}
	if n < 2 || ppy == 0 {
		return
	}
	rf := RiskFreeAnnualPct / 100 / ppy
	excess := make([]float64, n)
	var downside float64
	for i, x := range rets {
		excess[i] = x - rf
		if excess[i] < 0 {
			downside += excess[i] * excess[i]
		}
	}
	var mean, sd float64
	mean, sd, skew, kurt = moments(excess)
	if sd > 0 {
		sr = mean / sd
		r.Metrics.Sharpe = sr * math.Sqrt(ppy)
	}
	if dd := math.Sqrt(downside / float64(n)); dd > 0 {
		r.Metrics.Sortino = mean / dd * math.Sqrt(ppy)
	}
	growth := r.Curve[len(r.Curve)-1].Equity.Div(initial).InexactFloat64()
	if growth > 0 {
		r.Metrics.AnnualReturnPct = (math.Pow(growth, ppy/float64(n)) - 1) * 100
	}
	if mdd := r.Metrics.MaxDrawdownPct.InexactFloat64(); mdd > 0 {
		r.Metrics.Calmar = r.Metrics.AnnualReturnPct / mdd
	}
	pnl := make([]float64, len(r.TradePnL))
	for i, v := range r.TradePnL {
		pnl[i] = v.InexactFloat64()
	}
	if m, s, _, _ := moments(pnl); s > 0 {
		r.Metrics.SQN = math.Sqrt(float64(len(pnl))) * m / s
	}
	return sr, n, skew, kurt
}

func normCDF(x float64) float64 { return 0.5 * (1 + math.Erf(x/math.Sqrt2)) }
func normInv(p float64) float64 { return math.Sqrt2 * math.Erfinv(2*p-1) }

// DeflatedSharpe is the probability that the true Sharpe ratio is positive once the number of
// configurations tried is accounted for (Bailey and Lopez de Prado). sr is the per-bar Sharpe ratio
// of the selected run and n the number of its returns. Shortcut: the variance of the Sharpe
// estimates across trials is taken from the estimator's own standard error, because only the
// selected run is stored; upgrade it when every trial's Sharpe is kept.
func DeflatedSharpe(sr float64, n int, skew, kurt float64, trials int) float64 {
	if n < 3 {
		return 0
	}
	const euler = 0.5772156649015329
	sr0 := 0.0
	if trials > 1 {
		sd := math.Sqrt(1 / float64(n-1))
		t := float64(trials)
		sr0 = sd * ((1-euler)*normInv(1-1/t) + euler*normInv(1-1/(t*math.E)))
	}
	denom := 1 - skew*sr + (kurt-1)/4*sr*sr
	if denom <= 0 {
		return 0
	}
	return normCDF((sr - sr0) * math.Sqrt(float64(n-1)) / math.Sqrt(denom))
}

// splitMix is a small deterministic generator: the bootstrap must give the same answer on every run
// and every Go version, which the standard generators do not promise across releases.
type splitMix uint64

func (s *splitMix) next() uint64 {
	*s += 0x9e3779b97f4a7c15
	z := uint64(*s)
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// intn returns a value in [0, n). n is a positive slice length, far below 2^63.
func (s *splitMix) intn(n int) int {
	return int(s.next() % uint64(n)) //nolint:gosec // n is a positive slice length, far below 2^63
}

func bootstrap(pnl []d.Decimal, initial d.Decimal, seed uint64) *MonteCarlo {
	n := len(pnl)
	if n < 2 {
		return nil
	}
	block := min(bootstrapBlock, n)
	rng := splitMix(seed)
	totals := make([]float64, 0, bootstrapRuns)
	losses := 0
	for range bootstrapRuns {
		sum := Zero
		for taken := 0; taken < n; {
			start := rng.intn(n - block + 1)
			for j := 0; j < block && taken < n; j++ {
				sum = sum.Add(pnl[start+j])
				taken++
			}
		}
		pct := sum.Div(initial).Mul(D("100")).InexactFloat64()
		totals = append(totals, pct)
		if pct < 0 {
			losses++
		}
	}
	slices.Sort(totals)
	at := func(q float64) float64 { return totals[int(q*float64(len(totals)-1))] }
	return &MonteCarlo{
		Runs: bootstrapRuns, Block: block, P5Pct: at(0.05), P50Pct: at(0.5), P95Pct: at(0.95),
		ProbLoss: float64(losses) / float64(bootstrapRuns), SampleOf: n, Seed: seed,
		Technique: "block bootstrap of closed-trade P&L",
	}
}

func seedOf(parts ...string) uint64 {
	h := fnv.New64a()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{0})
	}
	return h.Sum64()
}

// robustness runs the overfitting checks for a finished run.
func robustness(
	ctx context.Context, ds Dataset, s Strategy, p Policy, partition string, r Result, trials int,
	sr float64, n int, skew, kurt float64,
) *Robustness {
	out := &Robustness{Trials: max(trials, 1), LowTrades: r.Metrics.Trades < MinTrades}
	out.DeflatedSharpe = DeflatedSharpe(sr, n, skew, kurt, out.Trials)
	out.MonteCarlo = bootstrap(r.TradePnL, Initial(), seedOf(ds.Hash, s.ID, s.Name, partition))
	stressed := p
	stressed.CostBPS = min(p.CostBPS*2, 100)
	stressed.SlippageBPS = min(p.SlippageBPS*2, 100)
	if stress, err := runBacktest(ctx, ds, s, stressed, partition); err == nil {
		v := stress.Metrics.ReturnPct
		out.StressReturn = &v
		out.StressNote = "Costs and slippage doubled (capped at 100 bps)."
	} else {
		out.StressNote = "Cost stress could not run: " + err.Error()
	}
	return out
}

// Initial is the starting capital of every backtest.
func Initial() d.Decimal { return D("100000") }
