package trading

import (
	"testing"
	"time"
)

func TestOperationalMetricsReportAgesRatesAndSlippage(t *testing.T) {
	now := time.Date(2026, 11, 4, 15, 0, 0, 0, time.UTC)
	a := armed()
	if m := OperationalMetrics(a, now); m.HeartbeatAgeSeconds != -1 || m.ReconciledAgeSeconds != -1 {
		t.Fatalf("an account that never reported has no age: %+v", m)
	}
	a.Heartbeat = now.Add(-20 * time.Second)
	a.ReconciledAt = now.Add(-90 * time.Second)
	// A buy signalled at 100 with 5 bps expected slippage has a limit of 100.05; it filled at 100.10.
	a.Orders = []Order{
		{
			ID:      "a",
			Symbol:  "AAPL",
			Side:    "buy",
			Status:  "filled",
			Limit:   D("100.05"),
			Average: D("100.10"),
			Filled:  D("10"),
			Qty:     D("10"),
		},
		{ID: "b", Symbol: "AAPL", Side: "buy", Status: "unknown", Limit: D("100"), Qty: D("5")},
		{ID: "c", Symbol: "AAPL", Side: "sell", Status: "dispatching", Limit: D("100"), Qty: D("5")},
	}
	a.RiskDecisions = []RiskDecision{{Key: "x"}}
	m := OperationalMetrics(a, now)
	if m.HeartbeatAgeSeconds != 20 || m.ReconciledAgeSeconds != 90 {
		t.Fatalf("ages: %+v", m)
	}
	if m.UnknownOrders != 1 || m.OpenOrdersWithoutFinal != 2 || m.FilledOrders != 1 {
		t.Fatalf("order counts: %+v", m)
	}
	if m.RejectedRate != 0.25 {
		t.Fatalf("one rejection against three orders is 25%%, got %v", m.RejectedRate)
	}
	// 100.10 against the 100 signal is 10 bps of slippage versus the 5 bps expected.
	if m.ObservedSlippageBPS < 9.9 || m.ObservedSlippageBPS > 10.1 || m.ExpectedSlippageBPS != 5 {
		t.Fatalf("slippage: observed %v expected %d", m.ObservedSlippageBPS, m.ExpectedSlippageBPS)
	}
}
