package trading

import (
	"strings"

	d "github.com/shopspring/decimal"
)

// The first live period is deliberately small. These ceilings are constants, not settings: raising them
// is a code change that goes through review and an ADR.
const (
	// FirstPeriodMaxOrderValue is the largest order value, in the account currency, a first-period
	// live deployment may configure.
	FirstPeriodMaxOrderValue = 100
	// FirstPeriodMaxDailyLossBPS is the largest daily loss limit, in basis points, for the same period.
	FirstPeriodMaxDailyLossBPS = 25
	// LiveAcknowledgement is the exact sentence the operator must set to accept real losses.
	LiveAcknowledgement = "I ACCEPT REAL LOSSES"
)

// LiveGate is the verdict on whether a live adapter may start. This repository has no live adapter and
// the verdict is Enabled == false unless every condition holds; any future adapter must call
// EvaluateLiveGate first and refuse to start without Enabled.
type LiveGate struct {
	Enabled       bool      `json:"enabled"`
	MaxOrderValue d.Decimal `json:"max_order_value"`
	MaxDailyLoss  int       `json:"max_daily_loss_bps"`
	Reasons       []string  `json:"reasons"`
}

// EvaluateLiveGate reads only the process environment, through get. The workspace, the risk policy, the
// MCP endpoint and every other setting an owner or an assistant can change at runtime are not inputs, so
// none of them can switch live trading on. role is "worker" or "web": live credentials belong in the
// worker only.
func EvaluateLiveGate(get func(string) string, role string) LiveGate {
	g := LiveGate{}
	deny := func(reason string) { g.Reasons = append(g.Reasons, reason) }
	if role != "worker" {
		deny("live credentials live only in the worker; the " + role + " service never enables live trading")
	}
	if get("LIVE_TRADING_ENABLED") != "true" {
		deny("LIVE_TRADING_ENABLED is not \"true\"")
	}
	if strings.TrimSpace(get("LIVE_TRADING_ACK")) != LiveAcknowledgement {
		deny("LIVE_TRADING_ACK does not carry the acknowledgement sentence")
	}
	value, err := d.NewFromString(get("LIVE_MAX_ORDER_VALUE"))
	switch {
	case err != nil || !value.IsPositive():
		deny("LIVE_MAX_ORDER_VALUE is not set to a positive number")
	case value.GreaterThan(d.NewFromInt(FirstPeriodMaxOrderValue)):
		deny("LIVE_MAX_ORDER_VALUE is above the first-period ceiling")
	default:
		g.MaxOrderValue = value
	}
	loss, ok := parseBPS(get("LIVE_MAX_DAILY_LOSS_BPS"))
	switch {
	case !ok || loss < 1:
		deny("LIVE_MAX_DAILY_LOSS_BPS is not set to a positive whole number")
	case loss > FirstPeriodMaxDailyLossBPS:
		deny("LIVE_MAX_DAILY_LOSS_BPS is above the first-period ceiling")
	default:
		g.MaxDailyLoss = loss
	}
	g.Enabled = len(g.Reasons) == 0
	return g
}

func parseBPS(s string) (int, bool) {
	n := 0
	if s == "" || len(s) > 6 {
		return 0, false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	return n, true
}

// WithinTolerance reports whether a paper result stays within a relative tolerance of what the backtest
// expected over the same period. It is one of the live-readiness criteria (ADR-0021); the comparison
// needs a real observation period, which only time can provide.
func WithinTolerance(expectedPct, observedPct, tolerance d.Decimal) bool {
	scale := expectedPct.Abs()
	if scale.LessThan(d.NewFromInt(1)) {
		scale = d.NewFromInt(1)
	}
	return observedPct.Sub(expectedPct).Abs().LessThanOrEqual(scale.Mul(tolerance))
}
