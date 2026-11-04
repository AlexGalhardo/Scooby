package trading

import (
	"strings"
	"testing"
	"time"

	d "github.com/shopspring/decimal"
)

func TestDefaultPreferencesAreValid(t *testing.T) {
	if err := DefaultPreferences().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPreferencesRejectBadWorkspaces(t *testing.T) {
	cases := map[string]func(*Preferences){
		"unknown account":    func(p *Preferences) { p.Account = "vault" },
		"unknown symbol":     func(p *Preferences) { p.Symbol = "XYZ" },
		"repeated watchlist": func(p *Preferences) { p.Watchlist = []string{"AAPL", "AAPL"} },
		"watchlist cap": func(p *Preferences) {
			p.Watchlist = nil
			for range MaxWatchlist + 1 {
				p.Watchlist = append(p.Watchlist, "AAPL")
			}
		},
		"alert without a price": func(p *Preferences) {
			p.Alerts = []PriceAlert{{ID: "a", Symbol: "AAPL", Direction: "above", Active: true}}
		},
		"alert direction": func(p *Preferences) {
			p.Alerts = []PriceAlert{{ID: "a", Symbol: "AAPL", Direction: "sideways", Price: D("1"), Active: true}}
		},
		"template without a stop": func(p *Preferences) {
			p.Templates = []OrderTemplate{{Key: "z", Label: "No stop", Side: "buy", CashPct: 5}}
		},
		"template over the cash cap": func(p *Preferences) {
			p.Templates = []OrderTemplate{
				{Key: "z", Label: "Big", Side: "buy", CashPct: 90, StopBPS: 100, TargetBPS: 200},
			}
		},
		"repeated hotkey": func(p *Preferences) {
			p.Templates = append(p.Templates, p.Templates[0])
		},
	}
	for name, mutate := range cases {
		p := DefaultPreferences()
		mutate(&p)
		if err := p.Validate(); err == nil {
			t.Fatalf("%s must be rejected", name)
		}
	}
}

func TestAlertsFireOnceWhenThePriceCrosses(t *testing.T) {
	p := DefaultPreferences()
	p.Alerts = []PriceAlert{
		{ID: "up", Symbol: "AAPL", Direction: "above", Price: D("200"), Active: true},
		{ID: "down", Symbol: "AAPL", Direction: "below", Price: D("150"), Active: true},
		{ID: "off", Symbol: "AAPL", Direction: "above", Price: D("100"), Active: false},
	}
	now := time.Date(2026, 11, 4, 15, 0, 0, 0, time.UTC)
	if got := p.Evaluate(map[string]d.Decimal{"AAPL": D("170")}, now); len(got) != 0 {
		t.Fatalf("a price between the levels fires nothing: %v", got)
	}
	got := p.Evaluate(map[string]d.Decimal{"AAPL": D("201")}, now)
	if len(got) != 1 || !strings.Contains(got[0], "above 200") {
		t.Fatalf("expected the upper alert, got %v", got)
	}
	if again := p.Evaluate(map[string]d.Decimal{"AAPL": D("205")}, now.Add(time.Minute)); len(again) != 0 {
		t.Fatalf("an alert fires once: %v", again)
	}
	if p.Alerts[0].Active || !p.Alerts[0].TriggeredAt.Equal(now) {
		t.Fatal("a fired alert is deactivated and timestamped")
	}
	if p.Alerts[2].TriggeredAt.IsZero() == false {
		t.Fatal("an inactive alert never fires")
	}
	if none := p.Evaluate(map[string]d.Decimal{}, now); len(none) != 0 {
		t.Fatal("an alert without a price stays quiet")
	}
}
