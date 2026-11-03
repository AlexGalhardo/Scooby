// Package engine runs independently of HTTP/browser sessions. The PostgreSQL
// session lock is held on a dedicated connection, including across broker I/O.
package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/AlexGalhardo/Scooby/internal/adapters"
	"github.com/AlexGalhardo/Scooby/internal/trading"
	d "github.com/shopspring/decimal"
	"log/slog"
	"strconv"
	"time"
)

type Engine struct {
	Store   trading.Store
	Broker  adapters.Alpaca
	Log     *slog.Logger
	Alert   func(context.Context, string)
	Version string
	Epoch   map[string]int64
}

func (e Engine) Run(ctx context.Context) error {
	if e.Store.DB.Dialect().Name() != "postgres" {
		return errors.New("worker requires PostgreSQL leadership locks")
	}
	conn, err := e.Store.DB.SQL().Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	var leader bool
	for !leader {
		if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(734662901)").Scan(&leader); err != nil {
			return err
		}
		if !leader {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(250 * time.Millisecond):
			}
		}
	}
	defer func() { _, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock(734662901)") }()
	if err := e.Store.Initialize(ctx); err != nil {
		return err
	}
	// A crashed job is never silently reported as completed. Safe to retry later.
	_, err = e.Store.DB.ExecContext(ctx, "UPDATE backtests SET status='queued' WHERE status='running'")
	if err != nil {
		return err
	}
	// Any in-flight send from a previous process must be reconciled, never resent.
	e.Epoch = map[string]int64{}
	for _, id := range []string{"b3", "us"} {
		if err := e.Store.Mutate(ctx, id, "engine", "boot", func(a *trading.State) error {
			a.LeaderEpoch++
			e.Epoch[id] = a.LeaderEpoch
			a.Generation++
			for j := range a.Orders {
				if a.Orders[j].Status == "dispatching" {
					a.Orders[j].Status = "unknown"
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	workerCtx, stopJobs := context.WithCancel(ctx)
	defer stopJobs()
	jobsDone := make(chan struct{})
	go func() {
		defer close(jobsDone)
		for {
			if workerCtx.Err() != nil {
				return
			}
			if err := e.Store.ProcessJob(workerCtx); err != nil && workerCtx.Err() == nil {
				e.Log.Error("backtest worker", "error", err)
			}
			select {
			case <-workerCtx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}()
	defer func() { stopJobs(); <-jobsDone }()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		if err := conn.PingContext(ctx); err != nil {
			return errors.New("leadership connection lost; worker stopped")
		}
		if err := e.Tick(ctx); err != nil {
			e.Log.Error("engine tick", "error", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
func (e Engine) Tick(ctx context.Context) error {
	strategies, err := e.Store.ActiveStrategies(ctx)
	if err != nil {
		return err
	}
	for _, id := range []string{"b3", "us"} {
		if err := e.mutate(ctx, id, "engine", "tick", func(a *trading.State) error { a.Heartbeat = time.Now().UTC(); return nil }); err != nil {
			return err
		}
	}
	for _, s := range strategies {
		if s.AccountID != "b3" {
			continue
		}
		var ds trading.Dataset
		if err := e.Store.Get(ctx, "datasets", s.DatasetID, &ds); err != nil {
			e.halt(ctx, "b3", "Dataset de replay indisponível")
			continue
		}
		if err := e.Replay(ctx, s, ds, 10); err != nil {
			e.halt(ctx, "b3", err.Error())
		}
	}
	if e.Broker.Key != "" {
		if err := e.Reconcile(ctx); err != nil {
			e.halt(ctx, "us", "Falha na reconciliação Alpaca: "+err.Error())
			return err
		}
		clock, err := e.Broker.Clock(ctx)
		if err != nil {
			e.halt(ctx, "us", "Calendário da corretora indisponível")
			return err
		}
		if !clock.Open || clock.NextClose.Sub(clock.Time) < 20*time.Minute {
			if err := e.mutate(ctx, "us", "engine", "session.cutoff", func(a *trading.State) error {
				for j := range a.Orders {
					o := &a.Orders[j]
					if o.Side == "buy" && o.Open() {
						if o.Status == "approved" {
							o.Status = "canceled"
							o.Reserved = trading.Zero
						} else {
							o.Status = "cancel_pending"
						}
					}
				}
				return nil
			}); err != nil {
				return err
			}
		}
		if clock.Open {
			for _, s := range strategies {
				if s.AccountID != "us" {
					continue
				}
				ds, err := e.Broker.Bars(ctx, s.Symbol, time.Now().Add(-7*24*time.Hour), clock.Time)
				if err != nil {
					e.halt(ctx, "us", "Feed IEX indisponível")
					continue
				}
				if err := e.SignalUS(ctx, s, ds, clock); err != nil {
					e.Log.Warn("signal skipped", "symbol", s.Symbol, "error", err)
				}
			}
		}
		if err := e.Dispatch(ctx); err != nil {
			return err
		}
	}
	return nil
}
func (e Engine) halt(ctx context.Context, id, reason string) {
	changed := false
	err := e.mutate(ctx, id, "engine", "halt", func(a *trading.State) error {
		changed = !a.Halted || a.Reason != reason
		if changed {
			a.Halt(reason)
		}
		return nil
	})
	if err != nil {
		e.Log.Error("cannot persist halt", "error", err)
	}
	if changed && e.Alert != nil {
		e.Alert(ctx, id+": "+reason)
	}
}
func (e Engine) Replay(ctx context.Context, s trading.Strategy, ds trading.Dataset, batch int) error {
	if ds.Symbol != s.Symbol {
		return errors.New("dataset de replay incompatível")
	}
	inst, _ := trading.FindInstrument(s.Symbol)
	return e.mutate(ctx, "b3", "engine", "tick", func(a *trading.State) error {
		for n := 0; n < batch; n++ {
			index := s.Slow + 1
			if v := a.Cursors[s.ID]; v != "" {
				parsed, err := strconv.Atoi(v)
				if err != nil {
					return err
				}
				index = parsed
			}
			if index >= len(ds.Candles) {
				return nil
			}
			c := ds.Candles[index]
			if a.Halted && len(a.Positions) == 0 {
				return nil
			}
			if err := c.Validate(); err != nil {
				return err
			}
			day := trading.MarketDay(inst.Market, c.Time)
			if a.Day != day {
				a.Day = day
				a.DayStart = a.Equity()
			}
			for j := range a.Orders {
				o := &a.Orders[j]
				if o.Symbol == s.Symbol && o.Open() && o.CreatedAt.Before(c.Time) {
					if trading.MarketDay(inst.Market, o.CreatedAt) != day {
						o.Status = "expired"
						o.Reserved = trading.Zero
						continue
					}
					if err := trading.Simulate(a, o, c); err != nil {
						return err
					}
				}
			}
			a.Mark(s.Symbol, c.Close)
			if p, ok := a.Positions[s.Symbol]; ok {
				price := trading.Zero
				if c.Low.LessThanOrEqual(p.Stop) {
					price = d.Min(p.Stop, c.Open)
				} else if c.High.GreaterThanOrEqual(p.Target) {
					price = p.Target
				} else if ds.ClosingAt(c.Time) || index == len(ds.Candles)-1 {
					price = c.Close
				}
				if price.IsPositive() {
					o, err := a.Reserve(s, "sell", s.ID+fmt.Sprint(index, "/exit"), price, c.Time)
					if err == nil {
						fillPrice := price.Mul(trading.D("1").Sub(trading.Fraction(a.Policy.SlippageBPS)))
						if err := a.ApplyFill(trading.Fill{ID: o.ID, OrderID: o.ID, Symbol: s.Symbol, Side: "sell", Qty: o.Qty, Price: fillPrice, Fees: o.Qty.Mul(fillPrice).Mul(trading.Fraction(a.Policy.CostBPS)), Time: c.Time}); err != nil {
							return err
						}
					}
				}
			}
			if !a.Halted && !ds.ClosingAt(c.Time) && index < len(ds.Candles)-1 {
				side := trading.Signal(s, ds.Candles[:index+1])
				if side != "hold" {
					key := s.ID + fmt.Sprint(index)
					_, err := a.Reserve(s, side, key, c.Close, c.Time)
					a.RecordRejection(key, s.Symbol, side, err, c.Time)
				}
			}
			a.Cursors[s.ID] = strconv.Itoa(index + 1)
			a.ReconciledAt = time.Now().UTC()
		}
		return nil
	})
}
func (e Engine) SignalUS(ctx context.Context, s trading.Strategy, ds trading.Dataset, clock adapters.Clock) error {
	if len(ds.Candles) < s.Slow+1 {
		return errors.New("warmup insuficiente")
	}
	last := ds.Candles[len(ds.Candles)-1]
	if clock.Time.Sub(last.Time.Add(5*time.Minute)) > 10*time.Minute {
		return errors.New("cotação obsoleta")
	}
	return e.mutate(ctx, "us", "engine", "tick", func(a *trading.State) error {
		if clock.Time.Sub(last.Time.Add(5*time.Minute)) > time.Duration(a.Policy.MaxAgeSeconds)*time.Second {
			a.Halt("Cotação obsoleta")
			return nil
		}
		day := trading.MarketDay("US", clock.Time)
		if a.Day != day {
			a.Day = day
			a.DayStart = a.Equity()
		}
		a.Mark(s.Symbol, last.Close)
		side := trading.Signal(s, ds.Candles)
		if p, ok := a.Positions[s.Symbol]; ok {
			if last.Close.LessThanOrEqual(p.Stop) || last.Close.GreaterThanOrEqual(p.Target) || clock.NextClose.Sub(clock.Time) < 15*time.Minute {
				side = "sell"
			}
		}
		if side == "hold" {
			return nil
		}
		if side == "buy" && clock.NextClose.Sub(clock.Time) < 20*time.Minute {
			return nil
		}
		key := s.ID + "/" + last.Time.Format(time.RFC3339) + "/" + side
		_, err := a.Reserve(s, side, key, last.Close, clock.Time)
		a.RecordRejection(key, s.Symbol, side, err, clock.Time)
		return nil
	})
}

// Reconcile recovers cumulative fills using cumulative notional differences. An
// unexplained broker position/order quarantines the account instead of adopting it.
func (e Engine) Reconcile(ctx context.Context) error {
	account, err := e.Broker.Account(ctx)
	if err != nil {
		return err
	}
	if account.ID == "" || account.Currency != "USD" || account.Status != "ACTIVE" || account.TradingBlocked {
		return errors.New("conta paper indisponível")
	}
	positions, err := e.Broker.Positions(ctx)
	if err != nil {
		return err
	}
	open, err := e.Broker.Orders(ctx)
	if err != nil {
		return err
	}
	var local trading.State
	if err := e.Store.Get(ctx, "trading_accounts", "us", &local); err != nil {
		return err
	}
	remote := map[string]adapters.BrokerOrder{}
	for _, o := range local.Orders {
		if o.Open() && o.Status != "approved" {
			b, err := e.Broker.Lookup(ctx, o.ID)
			if err != nil {
				return err
			}
			remote[o.ID] = b
		}
	}
	return e.mutate(ctx, "us", "engine", "tick", func(a *trading.State) error {
		if a.BrokerAccountID != "" && a.BrokerAccountID != account.ID {
			a.Halt("Credenciais apontam para outra conta Alpaca; migração administrativa necessária")
			return nil
		}
		a.BrokerAccountID = account.ID
		for _, b := range open {
			found := false
			for _, o := range a.Orders {
				if o.ID == b.ClientID {
					found = true
				}
			}
			if !found {
				a.Halt("Ordem externa não reconhecida")
				return nil
			}
		}
		for j := range a.Orders {
			o := &a.Orders[j]
			b, ok := remote[o.ID]
			if !ok {
				continue
			}
			if b.Symbol != o.Symbol || b.Side != o.Side || !b.Qty.Equal(o.Qty) {
				a.Halt("Ordem divergente")
				return nil
			}
			if b.Filled.LessThan(o.Filled) {
				a.Halt("Quantidade executada regrediu")
				return nil
			}
			if b.Filled.GreaterThan(o.Filled) {
				if b.Average == nil {
					return errors.New("fill sem preço")
				}
				delta := b.Filled.Sub(o.Filled)
				price := b.Average.Mul(b.Filled).Sub(o.Average.Mul(o.Filled)).Div(delta)
				f := trading.Fill{ID: b.ID + "/" + b.Filled.String(), OrderID: o.ID, Symbol: o.Symbol, Side: o.Side, Qty: delta, Price: price, Time: b.Updated}
				if err := a.ApplyFill(f); err != nil {
					return err
				}
			}
			o.BrokerID = b.ID
			o.UpdatedAt = b.Updated
			switch b.Status {
			case "filled", "canceled", "expired", "rejected":
				o.Status = b.Status
				o.Reserved = trading.Zero
			case "partially_filled":
				o.Status = b.Status
			case "pending_cancel":
				o.Status = "cancel_pending"
			default:
				if o.Status != "cancel_pending" {
					o.Status = "acknowledged"
				}
			}
		}
		seen := map[string]bool{}
		for _, p := range positions {
			expected, ok := a.Positions[p.Symbol]
			if !ok || !expected.Qty.Equal(p.Qty) || p.Qty.IsNegative() {
				a.Halt("Posição externa divergente: " + p.Symbol)
				return nil
			}
			a.Mark(p.Symbol, p.Price)
			seen[p.Symbol] = true
		}
		for symbol := range a.Positions {
			if !seen[symbol] {
				a.Halt("Posição local ausente na corretora")
				return nil
			}
		}
		// First contact is a balance baseline, never an invented paper balance.
		a.Cash = account.Cash
		if a.Initial.IsZero() {
			a.Initial = account.Equity
			a.DayStart = account.Equity
			a.Peak = account.Equity
		}
		a.ReconciledAt = time.Now().UTC()
		return nil
	})
}
func (e Engine) Dispatch(ctx context.Context) error {
	var a trading.State
	if err := e.Store.Get(ctx, "trading_accounts", "us", &a); err != nil {
		return err
	}
	for _, o := range a.Orders {
		if o.Status == "cancel_pending" {
			if o.BrokerID != "" {
				if err := e.Broker.Cancel(ctx, o.BrokerID); err != nil {
					return err
				}
			}
			continue
		}
		if o.Status != "approved" {
			continue
		}
		send := false
		if err := e.mutate(ctx, "us", "engine", "dispatch", func(current *trading.State) error {
			x := current.Order(o.ID)
			if x == nil || x.Status != "approved" {
				return nil
			}
			if x.Side == "buy" && (current.Halted || x.Generation != current.Generation) {
				x.Status = "canceled"
				x.Reserved = trading.Zero
				return nil
			}
			if time.Since(x.CreatedAt) > 2*time.Minute {
				x.Status = "expired"
				x.Reserved = trading.Zero
				return nil
			}
			x.Status = "dispatching"
			send = true
			return nil
		}); err != nil {
			return err
		}
		if !send {
			continue
		}
		b, err := e.Broker.Submit(ctx, o)
		persistErr := e.mutate(ctx, "us", "engine", "broker.response", func(current *trading.State) error {
			x := current.Order(o.ID)
			if x == nil {
				return sql.ErrNoRows
			}
			if err != nil {
				x.Status = "unknown"
				current.Halt("Ordem sem confirmação; reconciliação necessária")
			} else {
				x.BrokerID = b.ID
				x.Status = "acknowledged"
			}
			return nil
		})
		if persistErr != nil {
			return persistErr
		}
		if err != nil {
			return errors.New("broker request outcome unknown")
		}
	}
	return nil
}

func (e Engine) mutate(ctx context.Context, id, actor, action string, fn func(*trading.State) error) error {
	return e.Store.Mutate(ctx, id, actor, action, func(a *trading.State) error {
		if e.Epoch != nil && a.LeaderEpoch != e.Epoch[id] {
			return errors.New("stale worker fenced")
		}
		return fn(a)
	})
}
