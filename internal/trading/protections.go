package trading

import (
	"errors"
	"slices"
	"strconv"
	"time"

	d "github.com/shopspring/decimal"
)

// Close reasons recorded on every exit order. An exit without an explicit reason is a plain signal exit.
const (
	CloseStop    = "stop"
	CloseTarget  = "target"
	CloseSession = "session_close"
	CloseSignal  = "signal"
	CloseRisk    = "risk"
	CloseManual  = "manual"
)

const (
	haltDailyLoss     = "Daily loss limit"
	haltDrawdown      = "Equity drawdown limit"
	haltWaiting       = "Waiting for configuration and activation"
	unknownOrderAfter = 10 * time.Minute
)

func closeFor(side string) string {
	if side == "sell" {
		return CloseSignal
	}
	return ""
}

// CloseReasons lists the valid values of Order.CloseReason.
var CloseReasons = []string{CloseStop, CloseTarget, CloseSession, CloseSignal, CloseRisk, CloseManual}

// NewDay starts a trading day: it resets the loss baseline and lifts a daily-loss halt, which by
// design lasts until the next session. Every other halt stays until the owner clears it.
func (a *State) NewDay(day string) (resumed bool) {
	if a.Day == day {
		return false
	}
	a.Day = day
	a.DayStart = a.Equity()
	if a.Halted && a.Reason == haltDailyLoss {
		a.Halted = false
		a.Reason = ""
		a.Generation++
		return true
	}
	return false
}

// entryAllowed applies the protections that depend on history: the equity drawdown lock and the
// per-symbol locks (stop-loss streak and cooldown).
func (a *State) entryAllowed(symbol string, at time.Time) error {
	if dd := a.Policy.DrawdownBPS; dd > 0 && a.Peak.IsPositive() &&
		a.Equity().LessThanOrEqual(a.Peak.Mul(D("1").Sub(Fraction(dd)))) {
		a.Halt(haltDrawdown)
		return errors.New(haltDrawdown)
	}
	if until, ok := a.Locks[symbol]; ok && at.Before(until) {
		return errors.New("symbol locked")
	}
	return nil
}

// afterExit updates the streak and cooldown protections once a position is fully closed.
func (a *State) afterExit(symbol, reason string, pnl d.Decimal, at time.Time) {
	if a.Streaks == nil {
		a.Streaks = map[string]int{}
	}
	if a.Locks == nil {
		a.Locks = map[string]time.Time{}
	}
	if reason == CloseStop && pnl.IsNegative() {
		a.Streaks[symbol]++
	} else {
		a.Streaks[symbol] = 0
	}
	var until time.Time
	if n := a.Policy.StopStreak; n > 0 && a.Streaks[symbol] >= n {
		until = at.Add(time.Duration(a.Policy.StreakLockMins) * time.Minute)
		a.Streaks[symbol] = 0
	}
	if m := a.Policy.CooldownMins; m > 0 {
		if cool := at.Add(time.Duration(m) * time.Minute); cool.After(until) {
			until = cool
		}
	}
	if until.After(at) {
		a.Locks[symbol] = until
	}
}

// Watch returns the alerts that became true at now. Each alert fires once and stays quiet until its
// condition clears and comes back: an order stuck in `unknown`, or a halted account (kill switch,
// stale feed, divergence).
func (a *State) Watch(now time.Time) []string {
	active := map[string]string{}
	for _, o := range a.Orders {
		if o.Status == "unknown" && now.Sub(o.UpdatedAt) >= unknownOrderAfter {
			active["unknown:"+o.ID] = a.ID + ": order " + o.ID + " on " + o.Symbol + " unknown for " +
				strconv.Itoa(int(now.Sub(o.UpdatedAt).Minutes())) + " min"
		}
	}
	if a.Halted && a.Reason != "" && a.Reason != haltWaiting {
		active["halt:"+a.Reason] = a.ID + ": " + a.Reason
	}
	if a.Alerted == nil {
		a.Alerted = map[string]time.Time{}
	}
	var fresh []string
	for key, msg := range active {
		if _, seen := a.Alerted[key]; !seen {
			a.Alerted[key] = now
			fresh = append(fresh, msg)
		}
	}
	for key := range a.Alerted {
		if _, ok := active[key]; !ok {
			delete(a.Alerted, key)
		}
	}
	slices.Sort(fresh)
	return fresh
}
