package terminal

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/AlexGalhardo/Scooby/internal/adapters"
	"github.com/AlexGalhardo/Scooby/internal/trading"
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"
	"io"
	"net/http"
	"strings"
	"time"
)

type Controller struct {
	Store trading.Store
	App   *vukem.App
}

func New(app *vukem.App) (*Controller, error) {
	db, err := app.DB()
	return &Controller{trading.Store{DB: db}, app}, err
}
func decode(c *web.Ctx, out any) error {
	r := json.NewDecoder(io.LimitReader(c.Request().Body, 11<<20))
	r.DisallowUnknownFields()
	if err := r.Decode(out); err != nil {
		return errs.New(422, "invalid JSON or unknown fields")
	}
	var extra any
	if r.Decode(&extra) != io.EOF {
		return errs.New(422, "JSON must contain a single object")
	}
	return nil
}
func bad(err error) error {
	if err != nil {
		return errs.New(422, err.Error())
	}
	return nil
}
func accountID(id string) bool { return id == "b3" || id == "us" }
func (h *Controller) Page(section string) func(*web.Ctx) error {
	return func(c *web.Ctx) error { return inertia.Render(c, "Terminal", inertia.Props{"section": section}) }
}
func (h *Controller) Snapshot(c *web.Ctx) error {
	ctx := c.Context()
	if err := h.Store.Initialize(ctx); err != nil {
		return err
	}
	accounts, err := h.Store.List(ctx, "trading_accounts", 10)
	if err != nil {
		return err
	}
	states := []map[string]any{}
	for _, r := range accounts {
		var a trading.State
		if err := json.Unmarshal(r.Payload, &a); err != nil {
			return err
		}
		orders := a.Orders
		if len(orders) > 100 {
			orders = orders[len(orders)-100:]
		}
		fills := a.Fills
		if len(fills) > 100 {
			fills = fills[len(fills)-100:]
		}
		dd := trading.Zero
		if a.Peak.IsPositive() {
			dd = a.Peak.Sub(a.Equity()).Div(a.Peak).Mul(trading.D("100"))
		}
		unrealized := trading.Zero
		for _, p := range a.Positions {
			unrealized = unrealized.Add(p.Mark.Sub(p.Average).Mul(p.Qty))
		}
		states = append(states, map[string]any{"id": a.ID, "mode": a.Mode, "currency": a.Currency, "cash": a.Cash, "equity": a.Equity(), "realized": a.Realized, "unrealized": unrealized, "exposure": a.Exposure(), "drawdown": dd, "halted": a.Halted, "reason": a.Reason, "heartbeat": a.Heartbeat, "reconciled_at": a.ReconciledAt, "orders": orders, "positions": a.Positions, "fills": fills, "policy": a.Policy, "cursors": a.Cursors})
	}
	datasets, err := h.Store.List(ctx, "datasets", 100)
	if err != nil {
		return err
	}
	meta := []map[string]any{}
	for _, r := range datasets {
		var ds trading.Dataset
		if err := json.Unmarshal(r.Payload, &ds); err != nil {
			return err
		}
		var first, last time.Time
		if len(ds.Candles) > 0 {
			first = ds.Candles[0].Time
			last = ds.Candles[len(ds.Candles)-1].Time
		}
		meta = append(meta, map[string]any{"id": ds.ID, "symbol": ds.Symbol, "currency": ds.Currency, "source": ds.Source, "hash": ds.Hash, "license": ds.License, "bars": len(ds.Candles), "first": first, "last": last, "problems": trading.DatasetProblems(ds)})
	}
	strategies, err := h.Store.List(ctx, "strategies", 100)
	if err != nil {
		return err
	}
	jobs, err := h.Store.List(ctx, "backtests", 100)
	if err != nil {
		return err
	}
	for j := range jobs {
		var job trading.Job
		if err := json.Unmarshal(jobs[j].Payload, &job); err != nil {
			return err
		}
		if job.Result != nil {
			job.Result.Curve = nil
			job.Result.Fills = nil
			job.Result.Final = trading.State{}
		}
		jobs[j].Payload, _ = json.Marshal(job)
	}
	audit, err := h.Store.Audit(ctx)
	if err != nil {
		return err
	}
	return c.JSON(200, map[string]any{"accounts": states, "datasets": meta, "strategies": strategies, "backtests": jobs, "audit": audit, "instruments": trading.Instruments, "settings": map[string]any{"google_configured": h.App.Source.Get("SOCIAL_GOOGLE_CLIENT_ID") != "", "owner_configured": h.App.Source.Get("OWNER_GOOGLE_EMAIL") != "", "alpaca_configured": h.App.Source.Get("ALPACA_API_KEY") != "" && h.App.Source.Get("ALPACA_API_SECRET") != "", "mode": "paper_only", "feed": "IEX", "timeframe": 5, "engine_version": "scooby-1"}, "server_time": time.Now().UTC()})
}
func (h *Controller) Import(c *web.Ctx) error {
	var in struct {
		Symbol  string `json:"symbol"`
		Source  string `json:"source"`
		License string `json:"license"`
		CSV     string `json:"csv"`
	}
	if err := decode(c, &in); err != nil {
		return err
	}
	ds, err := trading.ImportCSV(strings.NewReader(in.CSV), in.Symbol, in.Source, in.License)
	if err != nil {
		return bad(err)
	}
	if err := h.Store.Create(c.Context(), "datasets", ds.ID, "ready", "owner", ds); err != nil {
		return err
	}
	return c.JSON(201, map[string]any{"id": ds.ID, "bars": len(ds.Candles)})
}
func (h *Controller) Fetch(c *web.Ctx) error {
	var in struct {
		Symbol string `json:"symbol"`
		Days   int    `json:"days"`
	}
	if err := decode(c, &in); err != nil {
		return err
	}
	if in.Days < 1 || in.Days > 30 {
		return errs.New(422, "period must be between 1 and 30 days")
	}
	broker := adapters.Alpaca{Key: h.App.Source.Get("ALPACA_API_KEY"), Secret: h.App.Source.Get("ALPACA_API_SECRET")}
	ctx, cancel := context.WithTimeout(c.Context(), 45*time.Second)
	defer cancel()
	ds, err := broker.Bars(ctx, in.Symbol, time.Now().Add(-time.Duration(in.Days)*24*time.Hour), time.Now().Add(-16*time.Minute))
	if err != nil {
		return errs.New(502, "Failed to fetch IEX history; check credentials, permissions and period")
	}
	if err := h.Store.Create(c.Context(), "datasets", ds.ID, "ready", "owner", ds); err != nil {
		return err
	}
	return c.JSON(201, map[string]any{"id": ds.ID, "bars": len(ds.Candles)})
}
func (h *Controller) Dataset(c *web.Ctx) error {
	var ds trading.Dataset
	if err := h.Store.Get(c.Context(), "datasets", c.Param("id"), &ds); err != nil {
		return errs.New(404, "Dataset not found")
	}
	if len(ds.Candles) > 1000 {
		ds.Candles = ds.Candles[len(ds.Candles)-1000:]
	}
	return c.JSON(200, ds)
}
func (h *Controller) Strategy(c *web.Ctx) error {
	var s trading.Strategy
	if err := decode(c, &s); err != nil {
		return err
	}
	if err := s.Validate(); err != nil {
		return bad(err)
	}
	if len(strings.TrimSpace(s.Name)) < 3 || len(s.Name) > 80 || !accountID(s.AccountID) {
		return errs.New(422, "Invalid name or account")
	}
	inst, _ := trading.FindInstrument(s.Symbol)
	if (s.AccountID == "b3") != (inst.Market == "B3") {
		return errs.New(422, "Symbol does not match the account")
	}
	if s.AccountID == "b3" {
		var ds trading.Dataset
		if err := h.Store.Get(c.Context(), "datasets", s.DatasetID, &ds); err != nil || ds.Symbol != s.Symbol || len(ds.Candles) < s.Slow+2 {
			return errs.New(422, "Select a dataset compatible with warmup")
		}
	}
	s.ID = trading.ID()
	s.Version = 1
	s.Enabled = false
	if err := h.Store.CreateStrategy(c.Context(), &s); err != nil {
		return err
	}
	return c.JSON(201, s)
}
func (h *Controller) Toggle(c *web.Ctx) error {
	var in struct {
		Enabled bool `json:"enabled"`
	}
	if err := decode(c, &in); err != nil {
		return err
	}
	var s trading.Strategy
	if err := h.Store.Get(c.Context(), "strategies", c.Param("id"), &s); err != nil {
		return errs.New(404, "Strategy not found")
	}
	if !in.Enabled {
		var a trading.State
		if err := h.Store.Get(c.Context(), "trading_accounts", s.AccountID, &a); err != nil {
			return err
		}
		if p := a.Positions[s.Symbol]; p.Qty.IsPositive() {
			return errs.New(422, "There is a position on the symbol. Block entries and keep the strategy until it is closed.")
		}
		for _, o := range a.Orders {
			if o.Symbol == s.Symbol && o.Open() {
				return errs.New(422, "There are pending orders on the symbol")
			}
		}
	}
	status := "disabled"
	if in.Enabled {
		status = "enabled"
	}
	if err := h.Store.ToggleStrategy(c.Context(), s.ID, in.Enabled); err != nil {
		return bad(err)
	}
	return c.JSON(200, map[string]string{"status": status})
}
func (h *Controller) Backtest(c *web.Ctx) error {
	var in struct {
		StrategyID string `json:"strategy_id"`
		DatasetID  string `json:"dataset_id"`
		Partition  string `json:"partition"`
	}
	if err := decode(c, &in); err != nil {
		return err
	}
	if in.Partition != "train" && in.Partition != "test" && in.Partition != "all" {
		return errs.New(422, "Invalid partition")
	}
	var s trading.Strategy
	var ds trading.Dataset
	if err := h.Store.Get(c.Context(), "strategies", in.StrategyID, &s); err != nil {
		return errs.New(422, "Invalid strategy")
	}
	if err := h.Store.Get(c.Context(), "datasets", in.DatasetID, &ds); err != nil || ds.Symbol != s.Symbol {
		return errs.New(422, "Incompatible dataset")
	}
	var a trading.State
	if err := h.Store.Get(c.Context(), "trading_accounts", s.AccountID, &a); err != nil {
		return err
	}
	job := trading.Job{ID: trading.ID(), DatasetID: ds.ID, Strategy: s, Policy: a.Policy, Partition: in.Partition, CreatedAt: time.Now().UTC(), CodeVersion: trading.BuildVersion}
	if err := h.Store.Create(c.Context(), "backtests", job.ID, "queued", "owner", job); err != nil {
		return err
	}
	return c.JSON(202, job)
}
func (h *Controller) Job(c *web.Ctx) error {
	var job trading.Job
	if err := h.Store.Get(c.Context(), "backtests", c.Param("id"), &job); err != nil {
		return errs.New(404, "Backtest not found")
	}
	if job.Result != nil {
		var result trading.Result
		if err := h.Store.Get(c.Context(), "artifacts", job.ID, &result); err != nil {
			return err
		}
		job.Result = &result
	}
	return c.JSON(200, job)
}
func (h *Controller) CancelJob(c *web.Ctx) error {
	if err := h.Store.Status(c.Context(), "backtests", c.Param("id"), "canceled", "owner"); err != nil {
		return errs.New(422, "Backtest already finished or does not exist")
	}
	return c.JSON(200, map[string]string{"status": "canceled"})
}
func (h *Controller) Policy(c *web.Ctx) error {
	id := c.Param("id")
	if !accountID(id) {
		return errs.New(404, "Invalid account")
	}
	var p trading.Policy
	if err := decode(c, &p); err != nil {
		return err
	}
	if err := p.Validate(); err != nil {
		return bad(err)
	}
	err := h.Store.Mutate(c.Context(), id, "owner", "risk.policy", func(a *trading.State) error {
		if !a.Halted {
			return errors.New("block the account before changing limits")
		}
		a.Policy = p
		a.Generation++
		return nil
	})
	if err != nil {
		return bad(err)
	}
	return c.JSON(200, p)
}
func (h *Controller) Control(c *web.Ctx) error {
	id := c.Param("id")
	if !accountID(id) {
		return errs.New(404, "Invalid account")
	}
	var in struct {
		Action  string `json:"action"`
		Confirm string `json:"confirm"`
	}
	if err := decode(c, &in); err != nil {
		return err
	}
	if in.Action == "reset" {
		if id != "b3" || in.Confirm != "RESET" {
			return errs.New(422, "reset allowed only on B3 replay with confirmation")
		}
		if err := h.Store.ResetReplay(c.Context()); err != nil {
			return bad(err)
		}
		return c.JSON(200, map[string]string{"status": "archived_and_reset"})
	}
	err := h.Store.Mutate(c.Context(), id, "owner", "control."+in.Action, func(a *trading.State) error {
		switch in.Action {
		case "halt":
			a.Halt("Block requested by the owner")
		case "resume":
			if in.Confirm != "PAPER" {
				return errors.New("PAPER confirmation required")
			}
			if a.Heartbeat.IsZero() || time.Since(a.Heartbeat) > 45*time.Second {
				return errors.New("worker is not healthy")
			}
			for _, o := range a.Orders {
				if o.Status == "unknown" || o.Status == "dispatching" || o.Status == "cancel_pending" {
					return errors.New("there are orders without a final state")
				}
			}
			if id == "us" && (a.ReconciledAt.IsZero() || time.Since(a.ReconciledAt) > 30*time.Second) {
				return errors.New("recent reconciliation required")
			}
			if a.DayStart.IsPositive() && a.Equity().LessThanOrEqual(a.DayStart.Mul(trading.D("1").Sub(trading.Fraction(a.Policy.DailyLossBPS)))) {
				return errors.New("daily loss limit reached")
			}
			a.Halted = false
			a.Reason = ""
			a.Generation++
		case "cancel":
			a.Halt("Cancellation requested")
			for j := range a.Orders {
				o := &a.Orders[j]
				if o.Open() {
					if o.Status == "approved" || id == "b3" {
						o.Status = "canceled"
						o.Reserved = trading.Zero
					} else {
						o.Status = "cancel_pending"
					}
				}
			}
		case "flatten":
			if in.Confirm != "ZERAR" {
				return errors.New("ZERAR confirmation required")
			}
			a.Halt("Flatten requested; monitor orders and positions")
			if a.ReconciledAt.IsZero() || time.Since(a.ReconciledAt) > 30*time.Second {
				return errors.New("recent reconciliation required; block and reconcile first")
			}
			for symbol, p := range a.Positions {
				s := trading.Strategy{Symbol: symbol}
				o, err := a.Reserve(s, "sell", trading.ID()+"/flatten", p.Mark, time.Now().UTC())
				if err != nil {
					return err
				}
				if id == "b3" {
					return errors.New("B3 replay: close through the strategy to preserve the historical clock")
				}
				_ = o
			}
		default:
			return errors.New("invalid action")
		}
		return nil
	})
	if err != nil {
		return bad(err)
	}
	return c.JSON(200, map[string]string{"status": "recorded"})
}
func (h *Controller) CancelOrder(c *web.Ctx) error {
	id := c.Param("account")
	if !accountID(id) {
		return errs.New(404, "Invalid account")
	}
	err := h.Store.Mutate(c.Context(), id, "owner", "order.cancel", func(a *trading.State) error {
		o := a.Order(c.Param("id"))
		if o == nil {
			return errors.New("order not found")
		}
		if o.Open() {
			if o.Status == "approved" || id == "b3" {
				o.Status = "canceled"
				o.Reserved = trading.Zero
			} else {
				o.Status = "cancel_pending"
			}
		}
		return nil
	})
	if err != nil {
		return bad(err)
	}
	return c.JSON(200, map[string]string{"status": "recorded"})
}
func (h *Controller) Ready(c *web.Ctx) error {
	if err := h.Store.DB.SQL().PingContext(c.Context()); err != nil {
		return c.JSON(503, map[string]string{"status": "unavailable"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}

func (h *Controller) Artifact(c *web.Ctx) error {
	var raw json.RawMessage
	if err := h.Store.Get(c.Context(), "artifacts", c.Param("id"), &raw); err != nil {
		return errs.New(404, "Artifact not found")
	}
	return c.JSON(200, raw)
}
