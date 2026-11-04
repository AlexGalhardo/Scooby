// Package trading implements deterministic, long-only simulation and risk.
package trading

import (
	"crypto/rand"
	"errors"
	"math"
	"slices"
	"sync"
	"time"

	d "github.com/shopspring/decimal"
)

var BuildVersion = "development"

func D(s string) d.Decimal {
	v, err := d.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return v
}
func ID() string { return rand.Text() }

var Zero = d.Zero

type Instrument struct {
	Symbol   string    `json:"symbol"`
	Market   string    `json:"market"`
	Currency string    `json:"currency"`
	Step     d.Decimal `json:"step"`
	Tick     d.Decimal `json:"tick"`
}

var Instruments = []Instrument{
	{"PETR4", "B3", "BRL", D("100"), D("0.01")},
	{"VALE3", "B3", "BRL", D("100"), D("0.01")},
	{"ITUB4", "B3", "BRL", D("100"), D("0.01")},
	{"AAPL", "US", "USD", D("1"), D("0.01")},
	{"NVDA", "US", "USD", D("1"), D("0.01")},
	{"TSLA", "US", "USD", D("1"), D("0.01")},
}

func FindInstrument(symbol string) (Instrument, error) {
	for _, i := range Instruments {
		if i.Symbol == symbol {
			return i, nil
		}
	}
	return Instrument{}, errors.New("symbol outside the watchlist")
}

type Candle struct {
	Time   time.Time `json:"time"`
	Open   d.Decimal `json:"open"`
	High   d.Decimal `json:"high"`
	Low    d.Decimal `json:"low"`
	Close  d.Decimal `json:"close"`
	Volume d.Decimal `json:"volume"`
}

func (c Candle) Validate() error {
	if c.Time.IsZero() || !c.Open.IsPositive() || !c.Close.IsPositive() || !c.Low.IsPositive() ||
		c.Volume.IsNegative() ||
		c.Low.GreaterThan(c.High) ||
		c.Open.GreaterThan(c.High) ||
		c.Close.GreaterThan(c.High) ||
		c.Open.LessThan(c.Low) ||
		c.Close.LessThan(c.Low) {
		return errors.New("invalid candle")
	}
	return nil
}

type Dataset struct {
	CalendarSource string               `json:"calendar_source"`
	SessionCloses  map[string]time.Time `json:"session_closes"`
	ID             string               `json:"id"`
	Symbol         string               `json:"symbol"`
	Source         string               `json:"source"`
	License        string               `json:"license"`
	Hash           string               `json:"hash"`
	Timeframe      int                  `json:"timeframe"`
	Currency       string               `json:"currency"`
	Synthetic      bool                 `json:"synthetic"`
	Candles        []Candle             `json:"candles"`
	CreatedAt      time.Time            `json:"created_at"`
}

// MinStopBPS is the smallest stop distance, in basis points, the risk engine accepts for an entry.
const MinStopBPS = 10

type Strategy struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Kind      string  `json:"kind"`
	Symbol    string  `json:"symbol"`
	DatasetID string  `json:"dataset_id"`
	AccountID string  `json:"account_id"`
	Version   int     `json:"version"`
	Fast      int     `json:"fast"`
	Slow      int     `json:"slow"`
	Threshold float64 `json:"threshold"`
	StopBPS   int     `json:"stop_bps"`
	TargetBPS int     `json:"target_bps"`
	Enabled   bool    `json:"enabled"`
}

func (s Strategy) Validate() error {
	if _, err := FindInstrument(s.Symbol); err != nil {
		return err
	}
	return s.validateParams()
}

func mean(cs []Candle) float64 {
	v := 0.
	for _, c := range cs {
		v += c.Close.InexactFloat64()
	}
	return v / float64(len(cs))
}

// Signal uses only the supplied CLOSED candles; execution occurs on a later bar.
func Signal(s Strategy, cs []Candle) string {
	n := len(cs)
	if n < s.Slow+1 {
		return "hold"
	}
	last := cs[n-1].Close.InexactFloat64()
	switch s.Kind {
	case "ma_cross":
		prev := mean(cs[n-1-s.Fast:n-1]) - mean(cs[n-1-s.Slow:n-1])
		curr := mean(cs[n-s.Fast:]) - mean(cs[n-s.Slow:])
		if prev <= 0 && curr > 0 {
			return "buy"
		}
		if prev >= 0 && curr < 0 {
			return "sell"
		}
	case "mean_reversion":
		window := cs[n-s.Slow:]
		m := mean(window)
		variance := 0.
		for _, c := range window {
			v := c.Close.InexactFloat64() - m
			variance += v * v
		}
		std := math.Sqrt(variance / float64(len(window)))
		if std < 1e-12 {
			return "hold"
		}
		z := (last - m) / std
		if z < -s.Threshold {
			return "buy"
		}
		if z >= 0 {
			return "sell"
		}
	case "breakout":
		high, low := cs[n-2].High, cs[n-2].Low
		for _, c := range cs[n-1-s.Slow : n-1] {
			high = d.Max(high, c.High)
			low = d.Min(low, c.Low)
		}
		if cs[n-1].Close.GreaterThan(high) {
			return "buy"
		}
		if cs[n-1].Close.LessThan(low) {
			return "sell"
		}
	case "opening_range":
		return openingRangeSignal(s, cs)
	case "vwap_reversion":
		return vwapSignal(s, cs)
	case "adx_trend":
		return adxSignal(s, cs)
	case "dca_grid":
		return dcaSignal(s, cs)
	}
	return "hold"
}

type Policy struct {
	RiskBPS       int `json:"risk_bps"`
	DailyLossBPS  int `json:"daily_loss_bps"`
	ExposureBPS   int `json:"exposure_bps"`
	MaxPositions  int `json:"max_positions"`
	CostBPS       int `json:"cost_bps"`
	SlippageBPS   int `json:"slippage_bps"`
	MaxAgeSeconds int `json:"max_age_seconds"`
	// Protections; zero disables each one, which keeps policies stored before they existed unchanged.
	StopStreak     int `json:"stop_streak"`      // consecutive stop-loss exits on a symbol before it is locked
	StreakLockMins int `json:"streak_lock_mins"` // length of that lock
	CooldownMins   int `json:"cooldown_mins"`    // pause on a symbol after any exit
	DrawdownBPS    int `json:"drawdown_bps"`     // equity drawdown from the peak that halts the account
}

func DefaultPolicy() Policy {
	return Policy{
		RiskBPS: 50, DailyLossBPS: 100, ExposureBPS: 5000, MaxPositions: 3, CostBPS: 5, SlippageBPS: 5,
		MaxAgeSeconds: 600, StopStreak: 3, StreakLockMins: 240, DrawdownBPS: 1000,
	}
}

func (p Policy) Validate() error {
	if p.RiskBPS < 1 || p.RiskBPS > 100 || p.DailyLossBPS < 1 || p.DailyLossBPS > 500 || p.ExposureBPS < 100 ||
		p.ExposureBPS > 10000 ||
		p.MaxPositions < 1 ||
		p.MaxPositions > 6 ||
		p.CostBPS < 0 ||
		p.CostBPS > 100 ||
		p.SlippageBPS < 1 ||
		p.SlippageBPS > 100 ||
		p.MaxAgeSeconds < 60 ||
		p.MaxAgeSeconds > 900 ||
		p.StopStreak < 0 || p.StopStreak > 10 ||
		p.StreakLockMins < 0 || p.StreakLockMins > 1440 || (p.StopStreak > 0 && p.StreakLockMins < 1) ||
		p.CooldownMins < 0 || p.CooldownMins > 1440 ||
		(p.DrawdownBPS != 0 && (p.DrawdownBPS < 100 || p.DrawdownBPS > 5000)) {
		return errors.New("invalid risk limits")
	}
	return nil
}
func Fraction(bps int) d.Decimal { return d.NewFromInt(int64(bps)).Div(D("10000")) }

type Position struct {
	Symbol  string    `json:"symbol"`
	Qty     d.Decimal `json:"qty"`
	Average d.Decimal `json:"average"`
	Mark    d.Decimal `json:"mark"`
	Stop    d.Decimal `json:"stop"`
	Target  d.Decimal `json:"target"`
	// Tranches counts the entry orders that filled into this position; LastEntry is the price of the
	// latest one. Both matter only to kinds that scale in.
	Tranches  int       `json:"tranches"`
	LastEntry d.Decimal `json:"last_entry"`
}
type Order struct {
	ID         string    `json:"id"`
	Key        string    `json:"key"`
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"`
	Qty        d.Decimal `json:"qty"`
	Filled     d.Decimal `json:"filled"`
	Average    d.Decimal `json:"average"`
	Limit      d.Decimal `json:"limit"`
	Stop       d.Decimal `json:"stop"`
	Target     d.Decimal `json:"target"`
	Reserved   d.Decimal `json:"reserved"`
	Status     string    `json:"status"`
	BrokerID   string    `json:"broker_id"`
	Reason     string    `json:"reason"`
	Close      string    `json:"close_reason"`
	StopBPS    int       `json:"stop_bps"`
	TargetBPS  int       `json:"target_bps"`
	Generation int64     `json:"generation"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (o Order) Open() bool {
	return slices.Contains(
		[]string{"approved", "dispatching", "unknown", "acknowledged", "partially_filled", "cancel_pending"},
		o.Status,
	)
}

type Fill struct {
	ID      string    `json:"id"`
	OrderID string    `json:"order_id"`
	Symbol  string    `json:"symbol"`
	Side    string    `json:"side"`
	Qty     d.Decimal `json:"qty"`
	Price   d.Decimal `json:"price"`
	Fees    d.Decimal `json:"fees"`
	Time    time.Time `json:"time"`
}
type State struct {
	BrokerAccountID string               `json:"broker_account_id"`
	RiskDecisions   []RiskDecision       `json:"risk_decisions"`
	LeaderEpoch     int64                `json:"leader_epoch"`
	ID              string               `json:"id"`
	Mode            string               `json:"mode"`
	Currency        string               `json:"currency"`
	Cash            d.Decimal            `json:"cash"`
	Initial         d.Decimal            `json:"initial"`
	DayStart        d.Decimal            `json:"day_start"`
	Day             string               `json:"day"`
	Realized        d.Decimal            `json:"realized"`
	Peak            d.Decimal            `json:"peak"`
	Policy          Policy               `json:"policy"`
	Halted          bool                 `json:"halted"`
	Reason          string               `json:"reason"`
	Generation      int64                `json:"generation"`
	Heartbeat       time.Time            `json:"heartbeat"`
	ReconciledAt    time.Time            `json:"reconciled_at"`
	Orders          []Order              `json:"orders"`
	Positions       map[string]Position  `json:"positions"`
	Fills           []Fill               `json:"fills"`
	Cursors         map[string]string    `json:"cursors"`
	Locks           map[string]time.Time `json:"locks"`
	Streaks         map[string]int       `json:"streaks"`
	Alerted         map[string]time.Time `json:"alerted"`
}

func NewState(id, currency, mode string) State {
	return State{
		ID:        id,
		Currency:  currency,
		Mode:      mode,
		Cash:      D("100000"),
		Initial:   D("100000"),
		DayStart:  D("100000"),
		Peak:      D("100000"),
		Policy:    DefaultPolicy(),
		Halted:    true,
		Reason:    haltWaiting,
		Orders:    []Order{},
		Fills:     []Fill{},
		Positions: map[string]Position{},
		Cursors:   map[string]string{},
	}
}

func (a State) Equity() d.Decimal {
	v := a.Cash
	for _, p := range a.Positions {
		v = v.Add(p.Qty.Mul(p.Mark))
	}
	return v
}

func (a State) Exposure() d.Decimal {
	v := Zero
	for _, p := range a.Positions {
		v = v.Add(p.Qty.Mul(p.Mark))
	}
	for _, o := range a.Orders {
		if o.Open() {
			v = v.Add(o.Reserved)
		}
	}
	return v
}
func (a *State) Halt(reason string) { a.Halted = true; a.Reason = reason; a.Generation++ }
func (a *State) Reserve(s Strategy, side, key string, price d.Decimal, at time.Time) (*Order, error) {
	for i := range a.Orders {
		if a.Orders[i].Key == key {
			return &a.Orders[i], nil
		}
	}
	i, err := FindInstrument(s.Symbol)
	if err != nil {
		return nil, err
	}
	if i.Currency != a.Currency {
		return nil, errors.New("currency mismatch")
	}
	if !price.IsPositive() {
		return nil, errors.New("invalid price")
	}
	for _, o := range a.Orders {
		if o.Open() && o.Symbol == s.Symbol {
			return nil, errors.New("pending order on symbol")
		}
	}
	var qty d.Decimal
	limit := price
	stop := price.Mul(D("1").Sub(Fraction(s.StopBPS))).RoundFloor(2)
	target := price.Mul(D("1").Add(Fraction(s.TargetBPS))).RoundFloor(2)
	reserve := Zero
	switch side {
	case "buy":
		if a.Halted {
			return nil, errors.New("account blocked")
		}
		if err := a.entryAllowed(s.Symbol, at); err != nil {
			return nil, err
		}
		if a.DayStart.IsPositive() &&
			a.Equity().LessThanOrEqual(a.DayStart.Mul(D("1").Sub(Fraction(a.Policy.DailyLossBPS)))) {
			a.Halt(haltDailyLoss)
			return nil, errors.New(a.Reason)
		}
		held := a.Positions[s.Symbol]
		spec, _ := FindKind(s.Kind)
		riskShare := d.NewFromInt(1)
		if held.Qty.IsPositive() {
			if !spec.Scales {
				return nil, errors.New("position already exists")
			}
			if held.Tranches >= s.Fast {
				return nil, errors.New("inventory cap reached")
			}
			step := Fraction(int(math.Round(s.Threshold * 100)))
			if price.GreaterThan(held.LastEntry.Mul(D("1").Sub(step))) {
				return nil, errors.New("tranche spacing not reached")
			}
		} else if len(a.Positions) >= a.Policy.MaxPositions {
			return nil, errors.New("position limit reached")
		}
		if spec.Scales {
			riskShare = d.NewFromInt(int64(max(s.Fast, 1)))
		}
		limit = price.Mul(D("1").Add(Fraction(a.Policy.SlippageBPS))).RoundCeil(2)
		if s.StopBPS < MinStopBPS || !stop.IsPositive() || !stop.LessThan(limit) {
			return nil, errors.New("stop required")
		}
		perUnit := limit.Sub(stop).Add(limit.Mul(Fraction(2*a.Policy.CostBPS + a.Policy.SlippageBPS)))
		remaining := d.Min(a.Cash, a.Equity().Mul(Fraction(a.Policy.ExposureBPS)).Sub(a.Exposure()))
		qty = d.Min(
			a.Equity().Mul(Fraction(a.Policy.RiskBPS)).Div(riskShare).Div(perUnit),
			remaining.Div(limit.Mul(D("1").Add(Fraction(a.Policy.CostBPS)))),
		)
		qty = qty.Div(i.Step).Floor().Mul(i.Step)
		reserve = qty.Mul(limit).Mul(D("1").Add(Fraction(a.Policy.CostBPS)))
	case "sell":
		qty = a.Positions[s.Symbol].Qty
		limit = price.Mul(D("1").Sub(Fraction(a.Policy.SlippageBPS))).RoundFloor(2)
	default:
		return nil, errors.New("invalid side")
	}
	if !qty.IsPositive() {
		return nil, errors.New("quantity below minimum")
	}
	o := Order{
		ID:         ID(),
		Key:        key,
		Symbol:     s.Symbol,
		Side:       side,
		Qty:        qty,
		Limit:      limit,
		Stop:       stop,
		Target:     target,
		Reserved:   reserve,
		Close:      closeFor(side),
		StopBPS:    s.StopBPS,
		TargetBPS:  s.TargetBPS,
		Status:     "approved",
		Generation: a.Generation,
		CreatedAt:  at,
		UpdatedAt:  at,
	}
	a.Orders = append(a.Orders, o)
	return &a.Orders[len(a.Orders)-1], nil
}

func (a *State) ApplyFill(f Fill) error {
	for _, old := range a.Fills {
		if old.ID == f.ID {
			return nil
		}
	}
	if f.ID == "" || !f.Qty.IsPositive() || !f.Price.IsPositive() || f.Fees.IsNegative() {
		return errors.New("invalid fill")
	}
	index := -1
	for i := range a.Orders {
		if a.Orders[i].ID == f.OrderID {
			index = i
			break
		}
	}
	if index < 0 {
		return errors.New("unknown order")
	}
	o := &a.Orders[index]
	if f.Side != o.Side || f.Symbol != o.Symbol || o.Filled.Add(f.Qty).GreaterThan(o.Qty) {
		return errors.New("fill exceeds intent")
	}
	value := f.Qty.Mul(f.Price)
	p := a.Positions[f.Symbol]
	if f.Side == "buy" {
		if value.Add(f.Fees).GreaterThan(a.Cash) {
			return errors.New("insufficient cash")
		}
		scaling := p.Qty.IsPositive()
		p.Average = p.Average.Mul(p.Qty).Add(value).Add(f.Fees).Div(p.Qty.Add(f.Qty))
		p.Qty = p.Qty.Add(f.Qty)
		p.Symbol = f.Symbol
		p.Mark = f.Price
		p.Stop = o.Stop
		p.Target = o.Target
		if scaling && o.StopBPS > 0 {
			// Adding to a position moves the exit levels to the new average cost.
			p.Stop = p.Average.Mul(D("1").Sub(Fraction(o.StopBPS))).RoundFloor(2)
			p.Target = p.Average.Mul(D("1").Add(Fraction(o.TargetBPS))).RoundFloor(2)
		}
		if o.Filled.IsZero() {
			p.Tranches++
		}
		p.LastEntry = f.Price
		a.Cash = a.Cash.Sub(value).Sub(f.Fees)
		a.Positions[f.Symbol] = p
	} else {
		if f.Qty.GreaterThan(p.Qty) {
			return errors.New("sell exceeds position")
		}
		a.Cash = a.Cash.Add(value).Sub(f.Fees)
		pnl := f.Price.Sub(p.Average).Mul(f.Qty).Sub(f.Fees)
		a.Realized = a.Realized.Add(pnl)
		p.Qty = p.Qty.Sub(f.Qty)
		if p.Qty.IsZero() {
			delete(a.Positions, f.Symbol)
			a.afterExit(f.Symbol, o.Close, pnl, f.Time)
		} else {
			p.Mark = f.Price
			a.Positions[f.Symbol] = p
		}
	}
	o.Average = o.Average.Mul(o.Filled).Add(value).Div(o.Filled.Add(f.Qty))
	o.Filled = o.Filled.Add(f.Qty)
	o.Reserved = o.Qty.Sub(o.Filled).Mul(o.Limit).Mul(D("1").Add(Fraction(a.Policy.CostBPS)))
	o.Status = "partially_filled"
	if o.Filled.Equal(o.Qty) {
		o.Status = "filled"
		o.Reserved = Zero
	}
	o.UpdatedAt = f.Time
	a.Fills = append(a.Fills, f)
	a.Peak = d.Max(a.Peak, a.Equity())
	return nil
}

func (a *State) Order(id string) *Order {
	for i := range a.Orders {
		if a.Orders[i].ID == id {
			return &a.Orders[i]
		}
	}
	return nil
}

func (a *State) Mark(symbol string, price d.Decimal) {
	if p, ok := a.Positions[symbol]; ok {
		p.Mark = price
		a.Positions[symbol] = p
		a.Peak = d.Max(a.Peak, a.Equity())
	}
}

var marketZones sync.Map // zone name to *time.Location; loading a zone reads the tz database

// MarketDay returns the calendar date of t in the market's time zone.
func MarketDay(market string, t time.Time) string {
	zone := "America/New_York"
	if market == "B3" {
		zone = "America/Sao_Paulo"
	}
	if cached, ok := marketZones.Load(zone); ok {
		if loc, ok := cached.(*time.Location); ok {
			return t.In(loc).Format("2006-01-02")
		}
	}
	loc, _ := time.LoadLocation(zone)
	marketZones.Store(zone, loc)
	return t.In(loc).Format("2006-01-02")
}

func Reject(o *Order, reason string) { o.Status = "rejected"; o.Reason = reason; o.Reserved = Zero }

type RiskDecision struct {
	Key      string    `json:"key"`
	Symbol   string    `json:"symbol"`
	Side     string    `json:"side"`
	Reason   string    `json:"reason"`
	Policy   Policy    `json:"policy"`
	Equity   d.Decimal `json:"equity"`
	Exposure d.Decimal `json:"exposure"`
	Time     time.Time `json:"time"`
}

func (a *State) RecordRejection(key, symbol, side string, err error, at time.Time) {
	if err == nil {
		return
	}
	for _, r := range a.RiskDecisions {
		if r.Key == key {
			return
		}
	}
	a.RiskDecisions = append(
		a.RiskDecisions,
		RiskDecision{key, symbol, side, err.Error(), a.Policy, a.Equity(), a.Exposure(), at},
	)
}
