package trading

import (
	"strings"
	"testing"
	"time"
)

func buyTemplate(pct int) OrderTemplate {
	return OrderTemplate{Key: "b", Label: "Buy", Side: "buy", CashPct: pct, StopBPS: 100, TargetBPS: 200}
}

func TestTicketNeverExceedsTheCashShareOrTheRiskBudget(t *testing.T) {
	now := time.Now()
	small := armed()
	o, err := small.Ticket(buyTemplate(1), "AAPL", D("100"), "k1", now)
	if err != nil {
		t.Fatal(err)
	}
	// 1% of 100000 cash is 1000, which buys at most 9 shares at about 100.1 with costs.
	if o.Qty.GreaterThan(D("9")) || !o.Qty.IsPositive() {
		t.Fatalf("the cash share must cap the size, got %s", o.Qty)
	}

	risky := armed()
	byStrategy, err := risky.Reserve(strategy(), "buy", "k0", D("100"), now)
	if err != nil {
		t.Fatal(err)
	}
	big := armed()
	ticket, err := big.Ticket(buyTemplate(50), "AAPL", D("100"), "k2", now)
	if err != nil {
		t.Fatal(err)
	}
	if ticket.Qty.GreaterThan(byStrategy.Qty) {
		t.Fatalf("the risk budget must still cap the size: ticket %s, risk-sized %s", ticket.Qty, byStrategy.Qty)
	}
	if !ticket.Stop.LessThan(ticket.Limit) {
		t.Fatal("a ticket entry carries its stop")
	}
}

func TestTicketBelowTheMinimumLeavesNothingBehind(t *testing.T) {
	a := armed()
	if _, err := a.Ticket(buyTemplate(1), "AAPL", D("5000"), "k", time.Now()); err == nil ||
		!strings.Contains(err.Error(), "minimum") {
		t.Fatalf("a ticket too small for one lot must be refused, got %v", err)
	}
	if len(a.Orders) != 0 || !a.Exposure().IsZero() {
		t.Fatal("a refused ticket must not reserve anything")
	}
}

func TestTicketGoesThroughTheRiskEngine(t *testing.T) {
	blocked := NewState("b", "USD", "simulation")
	if _, err := blocked.Ticket(buyTemplate(5), "AAPL", D("100"), "k", time.Now()); err == nil ||
		!strings.Contains(err.Error(), "account blocked") {
		t.Fatalf("a blocked account refuses tickets, got %v", err)
	}
	a := armed()
	a.Positions["AAPL"] = Position{Symbol: "AAPL", Qty: D("10"), Average: D("100"), Mark: D("100")}
	o, err := a.Ticket(OrderTemplate{Key: "x", Label: "Close", Side: "close"}, "AAPL", D("100"), "close", time.Now())
	if err != nil {
		t.Fatalf("a halted or armed account can always reduce risk: %v", err)
	}
	if o.Side != "sell" || o.Close != CloseManual || !o.Qty.Equal(D("10")) {
		t.Fatalf("a close ticket sells the whole position with a manual reason, got %+v", o)
	}
}
