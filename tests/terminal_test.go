package tests

import (
	"flag"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/vukem-trading/app/commands"
	"github.com/AlexGalhardo/vukem-trading/bootstrap"
	"github.com/AlexGalhardo/vukem-trading/internal/trading"
)

func terminalApp(t *testing.T, id, email string, verified bool) *vukemtest.App {
	t.Helper()
	v := "false"
	if verified {
		v = "true"
	}
	return vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{
			Env: map[string]string{
				"OWNER_GOOGLE_EMAIL":         "owner@example.com",
				"SOCIAL_FAKE_ID":             id,
				"SOCIAL_FAKE_EMAIL":          email,
				"SOCIAL_FAKE_EMAIL_VERIFIED": v,
			},
		},
	)
}

func signIn(t *testing.T, a *vukemtest.App) *vukemtest.Client {
	t.Helper()
	c := a.Client(t)
	r := c.Get("/auth/google/redirect")
	c.Get(r.Header("Location")).AssertRedirect("/")
	return c
}

func TestPrivateRoutesAndGoogleOnly(t *testing.T) {
	t.Parallel()
	a := terminalApp(t, "owner-1", "owner@example.com", true)
	c := a.Client(t)
	c.Get("/").AssertRedirect("/login")
	c.Get("/api/snapshot").AssertRedirect("/login")
	c.GetJSON("/api/snapshot").AssertUnauthorized()
	for _, path := range []string{"/register", "/forgot-password", "/posts"} {
		c.Get(path).AssertNotFound()
	}
	c.Get("/login").AssertOK().AssertInertia("Auth/Login")
	c.PostJSON("/login", map[string]any{}).AssertStatus(405)
	signIn(t, a).Get("/").AssertOK().AssertInertia("Terminal")
}

func TestGoogleGateAndReplay(t *testing.T) {
	for _, tc := range []struct {
		name, id, email string
		verified        bool
	}{{"foreign", "foreign", "intruder@example.com", true}, {"unverified", "owner", "owner@example.com", false}} {
		t.Run(tc.name, func(t *testing.T) {
			a := terminalApp(t, tc.id, tc.email, tc.verified)
			c := a.Client(t)
			r := c.Get("/auth/google/redirect")
			c.Get(r.Header("Location")).AssertForbidden()
			c.Get("/").AssertRedirect("/login")
			vukemtest.AssertDatabaseCount(t, a, "users", 0)
		})
	}
	a := terminalApp(t, "owner", "owner@example.com", true)
	c := a.Client(t)
	r := c.Get("/auth/google/redirect")
	c.Get(r.Header("Location")).AssertRedirect("/")
	c.Get(r.Header("Location")).AssertStatus(400)
	c.PostJSON("/logout", map[string]any{}).AssertRedirect("/login")
	c.Get("/").AssertRedirect("/login")
}

func TestTerminalPersistenceAndValidation(t *testing.T) {
	a := terminalApp(t, "owner", "owner@example.com", true)
	c := signIn(t, a).AcceptJSON()
	c.GetJSON("/api/snapshot").AssertOK().AssertJSONCount("accounts", 2).AssertJSONPath("settings.mode", "paper_only")
	db, err := a.DB()
	if err != nil {
		t.Fatal(err)
	}
	store := trading.Store{DB: db}
	c.PostJSON("/api/datasets/import", map[string]any{"symbol": "PETR4", "source": "licensed fixture", "license": "internal test", "csv": trading.SampleCSV()}).
		AssertCreated()
	vukemtest.AssertDatabaseCount(t, a, "datasets", 1)
	c.PostJSON("/api/datasets/import", map[string]any{"symbol": "PETR4", "source": "licensed fixture", "license": "test", "csv": "time,open,high,low,close,volume\n2026-10-26T10:00:00-03:00,NaN,30,29,30,100\n"}).
		AssertUnprocessable()
	c.PostJSON("/api/accounts/b3/control", map[string]any{"action": "resume", "confirm": "PAPER"}).AssertUnprocessable()
	c.PostJSON("/api/accounts/b3/control", map[string]any{"action": "halt"}).AssertOK()
	var state trading.State
	if err := store.Get(t.Context(), "trading_accounts", "b3", &state); err != nil {
		t.Fatal(err)
	}
	if !state.Halted || state.Generation != 1 {
		t.Fatalf("halt not persisted: %+v", state)
	}
	c.PostJSON("/api/accounts/foreign/control", map[string]any{"action": "halt"}).AssertNotFound()
	c.PostJSON("/api/accounts/b3/policy", map[string]any{"risk_bps": 90000}).AssertUnprocessable()
	c.PostJSON("/api/backtests", map[string]any{"strategy_id": "missing", "dataset_id": "missing", "partition": "test"}).
		AssertUnprocessable()
}

func TestOwnerRecoveryRevokesSessions(t *testing.T) {
	a := terminalApp(t, "owner-stable", "owner@example.com", true)
	c := signIn(t, a).AcceptJSON()
	c.GetJSON("/api/snapshot").AssertOK()
	command := &commands.OwnerReset{}
	fs := flag.NewFlagSet("recovery", flag.ContinueOnError)
	command.Flags(fs)
	if err := command.Run(t.Context(), a.App, nil); err == nil {
		t.Fatal("recovery without confirmation accepted")
	}
	if err := fs.Parse([]string{"--confirm", "RESET-OWNER", "--reason", "Owner recovery test"}); err != nil {
		t.Fatal(err)
	}
	if err := command.Run(t.Context(), a.App, nil); err != nil {
		t.Fatal(err)
	}
	c.GetJSON("/api/snapshot").AssertUnauthorized()
	signIn(t, a).Get("/").AssertOK()
	vukemtest.AssertDatabaseCount(t, a, "users", 1)
}

func TestStrategyVersionsActivationAndReplayArchive(t *testing.T) {
	app := terminalApp(t, "owner", "owner@example.com", true)
	c := signIn(t, app).AcceptJSON()
	c.GetJSON("/api/snapshot").AssertOK()
	input := map[string]any{
		"name":       "Versioned test",
		"kind":       "ma_cross",
		"symbol":     "AAPL",
		"account_id": "us",
		"fast":       2,
		"slow":       4,
		"threshold":  2,
		"stop_bps":   100,
		"target_bps": 200,
	}
	first := c.PostJSON("/api/strategies", input).AssertCreated().AssertJSONPath("version", 1)
	id, _ := first.JSONPath("id")
	second := c.PostJSON("/api/strategies", input).AssertCreated().AssertJSONPath("version", 2)
	secondID, _ := second.JSONPath("id")
	c.PostJSON("/api/strategies/"+id.(string)+"/toggle", map[string]any{"enabled": true}).AssertOK()
	c.PostJSON("/api/strategies/"+secondID.(string)+"/toggle", map[string]any{"enabled": true}).AssertUnprocessable()
	c.PostJSON("/api/accounts/b3/control", map[string]any{"action": "reset", "confirm": "RESET"}).AssertOK()
	vukemtest.AssertDatabaseCount(t, app, "artifacts", 1)
	c.PostJSON("/api/accounts/us/control", map[string]any{"action": "reset", "confirm": "RESET"}).AssertUnprocessable()
}
