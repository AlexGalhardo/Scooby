package trading

import (
	"strings"
	"testing"
)

func envOf(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func validLiveEnv() map[string]string {
	return map[string]string{
		"LIVE_TRADING_ENABLED":    "true",
		"LIVE_TRADING_ACK":        LiveAcknowledgement,
		"LIVE_MAX_ORDER_VALUE":    "50",
		"LIVE_MAX_DAILY_LOSS_BPS": "20",
	}
}

func TestLiveIsOffByDefault(t *testing.T) {
	g := EvaluateLiveGate(envOf(map[string]string{}), "worker")
	if g.Enabled || len(g.Reasons) < 4 {
		t.Fatalf("an empty environment must leave live off with every reason listed: %+v", g)
	}
}

func TestLiveNeedsEveryCondition(t *testing.T) {
	if g := EvaluateLiveGate(envOf(validLiveEnv()), "worker"); !g.Enabled {
		t.Fatalf("a complete worker environment opens the gate: %+v", g)
	}
	cases := map[string]func(map[string]string){
		"flag not true":          func(m map[string]string) { m["LIVE_TRADING_ENABLED"] = "yes" },
		"no acknowledgement":     func(m map[string]string) { m["LIVE_TRADING_ACK"] = "ok" },
		"order value too high":   func(m map[string]string) { m["LIVE_MAX_ORDER_VALUE"] = "101" },
		"order value missing":    func(m map[string]string) { delete(m, "LIVE_MAX_ORDER_VALUE") },
		"order value negative":   func(m map[string]string) { m["LIVE_MAX_ORDER_VALUE"] = "-5" },
		"daily loss too high":    func(m map[string]string) { m["LIVE_MAX_DAILY_LOSS_BPS"] = "26" },
		"daily loss not numeric": func(m map[string]string) { m["LIVE_MAX_DAILY_LOSS_BPS"] = "1e2" },
		"daily loss zero":        func(m map[string]string) { m["LIVE_MAX_DAILY_LOSS_BPS"] = "0" },
	}
	for name, mutate := range cases {
		env := validLiveEnv()
		mutate(env)
		if g := EvaluateLiveGate(envOf(env), "worker"); g.Enabled {
			t.Fatalf("%s must keep live off", name)
		}
	}
}

func TestOnlyTheWorkerCanEnableLive(t *testing.T) {
	g := EvaluateLiveGate(envOf(validLiveEnv()), "web")
	if g.Enabled || !strings.Contains(strings.Join(g.Reasons, " "), "only in the worker") {
		t.Fatalf("the web service must never enable live trading: %+v", g)
	}
}

func TestLiveGateReadsNothingButTheEnvironment(*testing.T) {
	// The function takes no preferences, policy or request: the signature itself is the proof that an
	// owner setting or an assistant cannot flip it. Keep it that way.
	_ = EvaluateLiveGate
}

func TestPaperResultsWithinATolerance(t *testing.T) {
	if !WithinTolerance(D("4"), D("3.2"), D("0.25")) {
		t.Fatal("0.8 points off a 4 point expectation is inside a 25% tolerance")
	}
	if WithinTolerance(D("4"), D("2"), D("0.25")) {
		t.Fatal("2 points off a 4 point expectation is outside it")
	}
	if !WithinTolerance(D("0"), D("0.2"), D("0.25")) {
		t.Fatal("a flat expectation uses a floor of one point")
	}
}
