package trading

import (
	"slices"
	"strings"
	"testing"
	"time"

	d "github.com/shopspring/decimal"
)

// roundTrip opens and fully closes a position, filling at the limit on entry and at exitPrice on exit.
func roundTrip(t *testing.T, a *State, key string, exitPrice string, reason string, at time.Time) {
	t.Helper()
	in, err := a.Reserve(strategy(), "buy", key+"/in", D("100"), at)
	if err != nil {
		t.Fatalf("%s entry: %v", key, err)
	}
	if err := a.ApplyFill(
		Fill{ID: key + "/in", OrderID: in.ID, Symbol: "AAPL", Side: "buy", Qty: in.Qty, Price: in.Limit, Time: at},
	); err != nil {
		t.Fatal(err)
	}
	out, err := a.Reserve(strategy(), "sell", key+"/out", D(exitPrice), at)
	if err != nil {
		t.Fatalf("%s exit: %v", key, err)
	}
	out.Close = reason
	if err := a.ApplyFill(
		Fill{
			ID:      key + "/out",
			OrderID: out.ID,
			Symbol:  "AAPL",
			Side:    "sell",
			Qty:     out.Qty,
			Price:   D(exitPrice),
			Time:    at,
		},
	); err != nil {
		t.Fatal(err)
	}
}

func TestNewDayLiftsOnlyTheDailyLossHalt(t *testing.T) {
	a := armed()
	a.Day = "2026-11-04"
	a.Halt(haltDailyLoss)
	generation := a.Generation
	if a.NewDay("2026-11-04") || !a.Halted {
		t.Fatal("the same day must not lift the halt")
	}
	if !a.NewDay("2026-11-05") || a.Halted || a.Generation <= generation {
		t.Fatalf("a new session must lift the daily-loss halt and advance the generation, halted=%v", a.Halted)
	}
	b := armed()
	b.Day = "2026-11-04"
	b.Halt("Order diverged")
	if b.NewDay("2026-11-05") || !b.Halted || b.Reason != "Order diverged" {
		t.Fatal("any other halt must stay until the owner clears it")
	}
}

func TestStopStreakLocksTheSymbolThenExpires(t *testing.T) {
	a := armed()
	a.Policy.StopStreak, a.Policy.StreakLockMins, a.Policy.DailyLossBPS = 2, 60, 500
	at := time.Date(2026, 11, 4, 15, 0, 0, 0, time.UTC)
	roundTrip(t, &a, "one", "98", CloseStop, at)
	roundTrip(t, &a, "two", "98", CloseStop, at.Add(time.Minute))
	if _, err := a.Reserve(strategy(), "buy", "three", D("100"), at.Add(2*time.Minute)); err == nil ||
		!strings.Contains(err.Error(), "symbol locked") {
		t.Fatalf("two stop-loss exits in a row must lock the symbol, got %v", err)
	}
	if _, err := a.Reserve(strategy(), "buy", "four", D("100"), at.Add(62*time.Minute)); err != nil {
		t.Fatalf("the lock must expire: %v", err)
	}
}

func TestATargetExitResetsTheStopStreak(t *testing.T) {
	a := armed()
	a.Policy.StopStreak, a.Policy.StreakLockMins, a.Policy.DailyLossBPS = 2, 60, 500
	at := time.Date(2026, 11, 4, 15, 0, 0, 0, time.UTC)
	roundTrip(t, &a, "one", "98", CloseStop, at)
	roundTrip(t, &a, "two", "103", CloseTarget, at.Add(time.Minute))
	roundTrip(t, &a, "three", "98", CloseStop, at.Add(2*time.Minute))
	if _, err := a.Reserve(strategy(), "buy", "four", D("100"), at.Add(3*time.Minute)); err != nil {
		t.Fatalf("a streak broken by a target exit must not lock: %v", err)
	}
}

func TestCooldownAppliesAfterAnyExit(t *testing.T) {
	a := armed()
	a.Policy.StopStreak, a.Policy.CooldownMins = 0, 30
	at := time.Date(2026, 11, 4, 15, 0, 0, 0, time.UTC)
	roundTrip(t, &a, "one", "103", CloseTarget, at)
	if _, err := a.Reserve(strategy(), "buy", "again", D("100"), at.Add(10*time.Minute)); err == nil {
		t.Fatal("a symbol in cooldown must reject entries")
	}
	if _, err := a.Reserve(strategy(), "buy", "later", D("100"), at.Add(31*time.Minute)); err != nil {
		t.Fatalf("cooldown must end: %v", err)
	}
}

func TestEquityDrawdownHaltsTheAccount(t *testing.T) {
	a := armed()
	a.Policy.DrawdownBPS = 1000
	a.Peak = D("100000")
	a.Cash = D("89000")
	a.DayStart = D("89000")
	if _, err := a.Reserve(strategy(), "buy", "dd", D("100"), time.Now()); err == nil ||
		!strings.Contains(err.Error(), haltDrawdown) {
		t.Fatalf("a 11%% drawdown must reject entries, got %v", err)
	}
	if !a.Halted || a.Reason != haltDrawdown {
		t.Fatal("the account must be halted with the drawdown reason")
	}
	a.NewDay("another-day")
	if !a.Halted {
		t.Fatal("a drawdown halt stays until the owner clears it")
	}
}

func TestEveryExitCarriesACloseReason(t *testing.T) {
	a := armed()
	at := time.Now()
	in, err := a.Reserve(strategy(), "buy", "in", D("100"), at)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.ApplyFill(
		Fill{ID: "in", OrderID: in.ID, Symbol: "AAPL", Side: "buy", Qty: in.Qty, Price: in.Limit, Time: at},
	); err != nil {
		t.Fatal(err)
	}
	out, err := a.Reserve(strategy(), "sell", "out", D("101"), at)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(CloseReasons, out.Close) {
		t.Fatalf("an exit order must carry a known close reason, got %q", out.Close)
	}
	if in.Close != "" {
		t.Fatal("an entry has no close reason")
	}
}

func TestWatchAlertsOncePerCondition(t *testing.T) {
	a := armed()
	now := time.Date(2026, 11, 4, 15, 0, 0, 0, time.UTC)
	a.Orders = append(
		a.Orders,
		Order{ID: "o1", Symbol: "AAPL", Status: "unknown", UpdatedAt: now.Add(-5 * time.Minute), Qty: d.NewFromInt(1)},
	)
	if got := a.Watch(now); len(got) != 0 {
		t.Fatalf("an order unknown for 5 minutes is not alarming yet: %v", got)
	}
	later := now.Add(6 * time.Minute)
	if got := a.Watch(later); len(got) != 1 || !strings.Contains(got[0], "unknown") {
		t.Fatalf("expected one unknown-order alert, got %v", got)
	}
	if got := a.Watch(later.Add(time.Minute)); len(got) != 0 {
		t.Fatalf("the same condition must not alert twice: %v", got)
	}
	a.Orders[0].Status = "filled"
	a.Watch(later.Add(2 * time.Minute))
	a.Orders[0].Status = "unknown"
	a.Orders[0].UpdatedAt = later.Add(-20 * time.Minute)
	if got := a.Watch(later.Add(3 * time.Minute)); len(got) != 1 {
		t.Fatalf("a condition that cleared and returned must alert again: %v", got)
	}

	a.Halt("Stale quote")
	if got := a.Watch(later.Add(4 * time.Minute)); len(got) != 1 || !strings.Contains(got[0], "Stale quote") {
		t.Fatalf("a halt must alert once, got %v", got)
	}
	if got := a.Watch(later.Add(5 * time.Minute)); len(got) != 0 {
		t.Fatalf("a halt alerts once: %v", got)
	}
}
