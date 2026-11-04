package trading

import (
	"time"

	d "github.com/shopspring/decimal"
)

// OpsMetrics are the operational signals of one account: whether the worker is alive, whether its view
// of the broker is fresh, how many orders are stuck, how often the risk engine says no, and how far fills
// land from what the policy expected.
type OpsMetrics struct {
	Account                string  `json:"account"`
	Halted                 bool    `json:"halted"`
	Reason                 string  `json:"reason"`
	HeartbeatAgeSeconds    float64 `json:"heartbeat_age_seconds"`  // -1 when the worker never reported
	ReconciledAgeSeconds   float64 `json:"reconciled_age_seconds"` // -1 when the account was never reconciled
	UnknownOrders          int     `json:"unknown_orders"`
	RejectedRate           float64 `json:"rejected_rate"` // risk rejections over rejections plus orders
	ObservedSlippageBPS    float64 `json:"observed_slippage_bps"`
	ExpectedSlippageBPS    int     `json:"expected_slippage_bps"`
	FilledOrders           int     `json:"filled_orders"`
	Generation             int64   `json:"generation"`
	AlertsOpen             int     `json:"alerts_open"`
	OpenOrdersWithoutFinal int     `json:"open_orders_without_final_state"`
}

func ageSeconds(at, now time.Time) float64 {
	if at.IsZero() || at.Year() < 2000 {
		return -1
	}
	return now.Sub(at).Seconds()
}

// OperationalMetrics derives the signals from the account state at now.
func OperationalMetrics(a State, now time.Time) OpsMetrics {
	m := OpsMetrics{
		Account: a.ID, Halted: a.Halted, Reason: a.Reason, Generation: a.Generation,
		HeartbeatAgeSeconds:  ageSeconds(a.Heartbeat, now),
		ReconciledAgeSeconds: ageSeconds(a.ReconciledAt, now),
		ExpectedSlippageBPS:  a.Policy.SlippageBPS,
		AlertsOpen:           len(a.Alerted),
	}
	slip := Fraction(a.Policy.SlippageBPS)
	var sum d.Decimal
	for _, o := range a.Orders {
		switch o.Status {
		case "unknown":
			m.UnknownOrders++
			m.OpenOrdersWithoutFinal++
		case "dispatching", "cancel_pending":
			m.OpenOrdersWithoutFinal++
		}
		if !o.Filled.IsPositive() || !o.Limit.IsPositive() {
			continue
		}
		// The limit is the signal price moved by the expected slippage, so the signal price is recoverable.
		var ref d.Decimal
		if o.Side == "buy" {
			ref = o.Limit.Div(D("1").Add(slip))
			sum = sum.Add(o.Average.Div(ref).Sub(D("1")).Mul(D("10000")))
		} else {
			ref = o.Limit.Div(D("1").Sub(slip))
			sum = sum.Add(D("1").Sub(o.Average.Div(ref)).Mul(D("10000")))
		}
		m.FilledOrders++
	}
	if m.FilledOrders > 0 {
		m.ObservedSlippageBPS = sum.Div(d.NewFromInt(int64(m.FilledOrders))).InexactFloat64()
	}
	if total := len(a.RiskDecisions) + len(a.Orders); total > 0 {
		m.RejectedRate = float64(len(a.RiskDecisions)) / float64(total)
	}
	return m
}
