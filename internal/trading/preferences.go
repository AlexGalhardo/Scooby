package trading

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	d "github.com/shopspring/decimal"
)

// Limits that keep the owner's saved workspace small and the checks cheap.
const (
	MaxWatchlist = 20
	MaxAlerts    = 50
	MaxTemplates = 6
	maxPctOfCash = 50
)

// PriceAlert fires once when the latest close crosses Price; the engine checks it on every tick.
type PriceAlert struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	Direction   string    `json:"direction"` // "above" or "below"
	Price       d.Decimal `json:"price"`
	Active      bool      `json:"active"`
	TriggeredAt time.Time `json:"triggered_at"`
}

// OrderTemplate is a one-key paper order: a side, a share of the available cash, and exit distances.
type OrderTemplate struct {
	Key       string `json:"key"`   // the hotkey, for example "b" or "shift+b"
	Label     string `json:"label"` // shown on the ticket
	Side      string `json:"side"`  // "buy" or "close"
	CashPct   int    `json:"cash_pct"`
	StopBPS   int    `json:"stop_bps"`
	TargetBPS int    `json:"target_bps"`
}

// Preferences is the owner's saved workspace. It is stored server-side so it follows the owner across
// browsers and the engine can read the alerts.
type Preferences struct {
	Account   string          `json:"account"`
	Symbol    string          `json:"symbol"`
	Watchlist []string        `json:"watchlist"`
	Alerts    []PriceAlert    `json:"alerts"`
	Templates []OrderTemplate `json:"templates"`
}

// DefaultPreferences is the workspace of a new owner.
func DefaultPreferences() Preferences {
	return Preferences{
		Account:   "b3",
		Symbol:    "PETR4",
		Watchlist: []string{"PETR4", "AAPL"},
		Alerts:    []PriceAlert{},
		Templates: []OrderTemplate{
			{Key: "b", Label: "Buy 5% of cash", Side: "buy", CashPct: 5, StopBPS: 100, TargetBPS: 200},
			{Key: "shift+b", Label: "Buy 10% of cash", Side: "buy", CashPct: 10, StopBPS: 100, TargetBPS: 200},
			{Key: "x", Label: "Close the position", Side: "close"},
		},
	}
}

// normalize turns absent lists into empty ones so the API never sends null for them.
func (p *Preferences) normalize() {
	if p.Watchlist == nil {
		p.Watchlist = []string{}
	}
	if p.Alerts == nil {
		p.Alerts = []PriceAlert{}
	}
	if p.Templates == nil {
		p.Templates = []OrderTemplate{}
	}
}

// Validate rejects anything that is not a known account, symbol or sensible limit.
func (p Preferences) Validate() error {
	if !slices.Contains([]string{"b3", "us", "research"}, p.Account) {
		return errors.New("unknown account")
	}
	if _, err := FindInstrument(p.Symbol); err != nil {
		return err
	}
	if len(p.Watchlist) > MaxWatchlist {
		return errors.New("the watchlist holds at most " + strconv.Itoa(MaxWatchlist) + " symbols")
	}
	seen := map[string]bool{}
	for _, s := range p.Watchlist {
		if _, err := FindInstrument(s); err != nil || seen[s] {
			return errors.New("the watchlist has an unknown or repeated symbol")
		}
		seen[s] = true
	}
	if len(p.Alerts) > MaxAlerts {
		return errors.New("too many alerts")
	}
	ids := map[string]bool{}
	for _, a := range p.Alerts {
		if _, err := FindInstrument(a.Symbol); err != nil || a.ID == "" || ids[a.ID] ||
			!slices.Contains([]string{"above", "below"}, a.Direction) || !a.Price.IsPositive() {
			return errors.New("invalid alert")
		}
		ids[a.ID] = true
	}
	if len(p.Templates) > MaxTemplates {
		return errors.New("too many order templates")
	}
	keys := map[string]bool{}
	for _, t := range p.Templates {
		if t.Key == "" || t.Label == "" || keys[t.Key] {
			return errors.New("invalid order template")
		}
		keys[t.Key] = true
		switch t.Side {
		case "buy":
			if t.CashPct < 1 || t.CashPct > maxPctOfCash || t.StopBPS < MinStopBPS || t.StopBPS > 2000 ||
				t.TargetBPS < 10 || t.TargetBPS > 5000 {
				return errors.New("invalid order template")
			}
		case "close":
		default:
			return errors.New("invalid order template")
		}
	}
	return nil
}

// Evaluate checks every active alert against the latest prices. It deactivates the alerts that fired
// and returns their messages, so each alert fires exactly once.
func (p *Preferences) Evaluate(prices map[string]d.Decimal, now time.Time) []string {
	var fired []string
	for i := range p.Alerts {
		a := &p.Alerts[i]
		price, ok := prices[a.Symbol]
		if !a.Active || !ok {
			continue
		}
		if (a.Direction == "above" && price.GreaterThanOrEqual(a.Price)) ||
			(a.Direction == "below" && price.LessThanOrEqual(a.Price)) {
			a.Active = false
			a.TriggeredAt = now
			fired = append(
				fired,
				"Alert: "+a.Symbol+" is "+a.Direction+" "+a.Price.String()+" (last "+price.String()+")",
			)
		}
	}
	return fired
}

// Preferences returns the saved workspace, or the defaults before the first save.
func (s Store) Preferences(ctx context.Context) (Preferences, error) {
	var body string
	err := s.DB.QueryRowContext(ctx, "SELECT payload FROM preferences WHERE id='owner'").Scan(&body)
	if errors.Is(err, sql.ErrNoRows) {
		return DefaultPreferences(), nil
	}
	if err != nil {
		return Preferences{}, err
	}
	p := DefaultPreferences()
	if err := json.Unmarshal([]byte(body), &p); err != nil {
		return Preferences{}, err
	}
	p.normalize()
	return p, nil
}

// SavePreferences validates and stores the workspace.
func (s Store) SavePreferences(ctx context.Context, p Preferences) error {
	p.normalize()
	if err := p.Validate(); err != nil {
		return err
	}
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(
		ctx,
		s.q(`INSERT INTO preferences (id,payload,version,updated_at,status) VALUES ('owner',?,1,?,'ready')
		ON CONFLICT (id) DO UPDATE SET payload=excluded.payload, version=preferences.version+1, updated_at=excluded.updated_at`),
		string(b),
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// LastPrices returns the latest close of each requested symbol, taken from the freshest dataset.
func (s Store) LastPrices(ctx context.Context, symbols ...string) (map[string]d.Decimal, error) {
	records, err := s.List(ctx, "datasets", 200)
	if err != nil {
		return nil, err
	}
	prices := map[string]d.Decimal{}
	latest := map[string]time.Time{}
	for _, r := range records {
		var ds Dataset
		if json.Unmarshal(r.Payload, &ds) != nil || len(ds.Candles) == 0 || !slices.Contains(symbols, ds.Symbol) {
			continue
		}
		last := ds.Candles[len(ds.Candles)-1]
		if last.Time.After(latest[ds.Symbol]) {
			latest[ds.Symbol] = last.Time
			prices[ds.Symbol] = last.Close
		}
	}
	return prices, nil
}

// UpdatePreferences reads the saved workspace under a row lock, applies fn and stores the result in
// one transaction, so the engine marking an alert as fired cannot overwrite an edit by the owner.
func (s Store) UpdatePreferences(ctx context.Context, fn func(*Preferences) error) error {
	if err := s.SavePreferencesIfMissing(ctx); err != nil {
		return err
	}
	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		q := "SELECT payload FROM preferences WHERE id='owner'"
		if tx.Dialect().Name() == "postgres" {
			q += " FOR UPDATE"
		}
		var body string
		if err := tx.QueryRowContext(ctx, q).Scan(&body); err != nil {
			return err
		}
		p := DefaultPreferences()
		if err := json.Unmarshal([]byte(body), &p); err != nil {
			return err
		}
		if err := fn(&p); err != nil {
			return err
		}
		p.normalize()
		if err := p.Validate(); err != nil {
			return err
		}
		b, err := json.Marshal(p)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, database.Rebind(tx.Dialect(),
			"UPDATE preferences SET payload=?, version=version+1, updated_at=? WHERE id='owner'"),
			string(b), time.Now().UTC().Format(time.RFC3339))
		return err
	})
}

// SavePreferencesIfMissing stores the defaults when the owner has not saved a workspace yet.
func (s Store) SavePreferencesIfMissing(ctx context.Context) error {
	b, err := json.Marshal(DefaultPreferences())
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(
		ctx,
		s.q(`INSERT INTO preferences (id,payload,version,updated_at,status) VALUES ('owner',?,0,?,'ready')
		ON CONFLICT (id) DO NOTHING`),
		string(b),
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}
