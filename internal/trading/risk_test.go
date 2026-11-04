package trading

import (
	"strings"
	"testing"
	"time"

	d "github.com/shopspring/decimal"
)

func armed() State {
	a := NewState("risk", "USD", "simulation")
	a.Halted = false
	a.Reason = ""
	return a
}

func TestEntryWithoutStopIsRejected(t *testing.T) {
	for _, bps := range []int{0, MinStopBPS - 1, -5} {
		a := armed()
		s := strategy()
		s.StopBPS = bps
		if _, err := a.Reserve(s, "buy", "no-stop", D("100"), time.Now()); err == nil ||
			!strings.Contains(err.Error(), "stop required") {
			t.Fatalf("stop %d bps: want a stop required rejection, got %v", bps, err)
		}
		if len(a.Orders) != 0 || !a.Exposure().IsZero() {
			t.Fatalf("stop %d bps: a rejected entry must not reserve anything", bps)
		}
	}
}

func TestEntryCarriesItsStop(t *testing.T) {
	a := armed()
	o, err := a.Reserve(strategy(), "buy", "with-stop", D("100"), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !o.Stop.IsPositive() || !o.Stop.LessThan(o.Limit) {
		t.Fatalf("stop %s must sit below the entry %s", o.Stop, o.Limit)
	}
}

func TestRiskPerTradeSizingInBasisPoints(t *testing.T) {
	for _, bps := range []int{25, 50, 100} {
		a := armed()
		a.Policy.RiskBPS = bps
		s := strategy()
		o, err := a.Reserve(s, "buy", "size", D("100"), time.Now())
		if err != nil {
			t.Fatalf("%d bps: %v", bps, err)
		}
		perUnit := o.Limit.Sub(o.Stop).Add(o.Limit.Mul(Fraction(2*a.Policy.CostBPS + a.Policy.SlippageBPS)))
		loss := o.Qty.Mul(perUnit)
		budget := a.Equity().Mul(Fraction(bps))
		if loss.GreaterThan(budget) {
			t.Fatalf("%d bps: loss at the stop %s exceeds the budget %s", bps, loss, budget)
		}
		step := d.NewFromInt(1)
		if budget.Sub(loss).GreaterThan(perUnit.Mul(step)) &&
			o.Reserved.LessThan(a.Equity().Mul(Fraction(a.Policy.ExposureBPS)).Sub(o.Limit)) {
			t.Fatalf("%d bps: sized more than one step below the budget (%s of %s)", bps, loss, budget)
		}
	}
}

func TestDefaultDenyRejectsEntriesButAllowsExits(t *testing.T) {
	a := NewState("deny", "USD", "simulation")
	if !a.Halted {
		t.Fatal("a new account must start blocked")
	}
	if _, err := a.Reserve(strategy(), "buy", "blocked", D("100"), time.Now()); err == nil ||
		!strings.Contains(err.Error(), "account blocked") {
		t.Fatalf("blocked account must reject entries with a reason, got %v", err)
	}
	a.Positions["AAPL"] = Position{Symbol: "AAPL", Qty: D("10"), Average: D("100"), Mark: D("100")}
	if _, err := a.Reserve(strategy(), "sell", "exit", D("100"), time.Now()); err != nil {
		t.Fatalf("a blocked account must still allow reducing risk: %v", err)
	}
}

func TestDailyLossKillSwitchBlocksEntriesNotExits(t *testing.T) {
	a := armed()
	a.Positions["AAPL"] = Position{Symbol: "AAPL", Qty: D("1000"), Average: D("100"), Mark: D("100")}
	a.Cash = D("0")
	a.DayStart = a.Equity()
	// A fall of more than DailyLossBPS (100 bps) in the open position.
	a.Positions["AAPL"] = Position{Symbol: "AAPL", Qty: D("1000"), Average: D("100"), Mark: D("98")}

	if _, err := a.Reserve(strategy(), "buy", "after-loss", D("98"), time.Now()); err == nil {
		t.Fatal("entries must be blocked once the daily loss limit is hit")
	}
	if !a.Halted || !strings.Contains(a.Reason, "Daily loss") {
		t.Fatalf("the account must be halted with a reason, got halted=%v reason=%q", a.Halted, a.Reason)
	}
	generation := a.Generation
	if generation == 0 {
		t.Fatal("halting must advance the generation so in-flight work is fenced")
	}
	if _, err := a.Reserve(strategy(), "sell", "reduce", D("98"), time.Now()); err != nil {
		t.Fatalf("exits must stay allowed after the kill switch: %v", err)
	}
}
