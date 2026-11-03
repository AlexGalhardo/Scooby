package trading

import (
	"context"
	"errors"
	"fmt"
	d "github.com/shopspring/decimal"
	"time"
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
	CalendarSource string   `json:"calendar_source"`
	Metrics        Metrics  `json:"metrics"`
	Curve          []Point  `json:"curve"`
	Fills          []Fill   `json:"fills"`
	DatasetHash    string   `json:"dataset_hash"`
	Strategy       Strategy `json:"strategy"`
	Policy         Policy   `json:"policy"`
	Warnings       []string `json:"warnings"`
	Final          State    `json:"-"`
	EngineVersion  string   `json:"engine_version"`
	Partition      string   `json:"partition"`
}

func Backtest(ctx context.Context, ds Dataset, s Strategy, policy Policy, partition string) (Result, error) {
	r := Result{CalendarSource: ds.CalendarSource, Curve: []Point{}, Warnings: []string{"Simulação: OHLC não revela sequência intrabar; stop tem prioridade sobre alvo.", "Custos em bps; impostos não incluídos. Sem garantia de rentabilidade.", "Benchmark bruto buy-and-hold; posições remanescentes são encerradas no final da amostra."}, DatasetHash: ds.Hash, Strategy: s, Policy: policy, EngineVersion: BuildVersion, Partition: partition}
	if err := s.Validate(); err != nil {
		return r, err
	}
	if err := policy.Validate(); err != nil {
		return r, err
	}
	if ds.Symbol != s.Symbol {
		return r, errors.New("dataset não corresponde ao ativo")
	}
	cs := ds.Candles
	split := len(cs) * 7 / 10
	switch partition {
	case "train":
		cs = cs[:split]
	case "test":
		cs = cs[split:]
	case "all":
		r.Warnings = append(r.Warnings, "Exploratório: usa o período completo, não é teste fora da amostra.")
	default:
		return r, errors.New("partição deve ser train, test ou all")
	}
	if len(cs) < s.Slow+2 {
		return r, errors.New("histórico insuficiente para warmup e avaliação")
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
			o := a.Order(pending)
			if o != nil && MarketDay(i.Market, o.CreatedAt) == day {
				if err := Simulate(&a, o, c); err != nil {
					return r, err
				}
			} else if o != nil {
				o.Status = "expired"
				o.Reserved = Zero
			}
			pending = ""
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
				o := Order{ID: ID(), Key: fmt.Sprint(idx, "/exit"), Symbol: s.Symbol, Side: "sell", Qty: p.Qty, Status: "approved", CreatedAt: c.Time}
				a.Orders = append(a.Orders, o)
				price := exit.Mul(D("1").Sub(Fraction(policy.SlippageBPS)))
				if err := a.ApplyFill(Fill{ID: o.ID, OrderID: o.ID, Symbol: s.Symbol, Side: "sell", Qty: p.Qty, Price: price, Fees: p.Qty.Mul(price).Mul(Fraction(policy.CostBPS)), Time: c.Time}); err != nil {
					return r, err
				}
			}
		}
		if cycleOpen && !a.Positions[s.Symbol].Qty.IsPositive() {
			gain := a.Realized.Sub(cycleStart)
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
