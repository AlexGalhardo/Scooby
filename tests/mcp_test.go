package tests

import (
	"regexp"
	"slices"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/vukem-trading/app/controllers/terminal"
	"github.com/AlexGalhardo/vukem-trading/bootstrap"
)

func mcpApp(t *testing.T, token string) *vukemtest.App {
	t.Helper()
	env := map[string]string{
		"OWNER_GOOGLE_EMAIL": "owner@example.com", "SOCIAL_FAKE_ID": "owner",
		"SOCIAL_FAKE_EMAIL": "owner@example.com", "SOCIAL_FAKE_EMAIL_VERIFIED": "true",
	}
	if token != "" {
		env["MCP_TOKEN"] = token
	}
	return vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})
}

func rpc(method string, params any) map[string]any {
	return map[string]any{"jsonrpc": "2.0", "id": 1, "method": method, "params": params}
}

func TestMCPIsOffWithoutAToken(t *testing.T) {
	app := mcpApp(t, "")
	app.Client(t).PostJSON("/mcp", rpc("tools/list", map[string]any{})).AssertNotFound()
}

func TestMCPRejectsMissingOrWrongTokens(t *testing.T) {
	app := mcpApp(t, "correct-horse-battery-staple-0123456789")
	c := app.Client(t)
	c.PostJSON("/mcp", rpc("tools/list", map[string]any{})).AssertUnauthorized()
	app.Client(t).
		WithHeader("Authorization", "Bearer wrong").
		PostJSON("/mcp", rpc("tools/list", map[string]any{})).
		AssertUnauthorized()
	app.Client(t).
		WithHeader("Authorization", "correct-horse-battery-staple-0123456789").
		PostJSON("/mcp", rpc("tools/list", map[string]any{})).
		AssertUnauthorized()
}

func TestMCPOffersOnlyReadOnlyTools(t *testing.T) {
	app := mcpApp(t, "correct-horse-battery-staple-0123456789")
	c := app.Client(t).WithToken("correct-horse-battery-staple-0123456789")
	c.PostJSON("/mcp", rpc("initialize", map[string]any{})).
		AssertOK().
		AssertJSONPath("result.serverInfo.name", "vukem-trading")
	res := c.PostJSON("/mcp", rpc("tools/list", map[string]any{})).AssertOK()
	raw, _ := res.JSONPath("result.tools")
	tools, _ := raw.([]any)
	var names []string
	for _, item := range tools {
		if m, ok := item.(map[string]any); ok {
			names = append(names, m["name"].(string))
		}
	}
	want := []string{
		"get_backtest_report",
		"get_ops_metrics",
		"list_datasets",
		"list_instruments",
		"list_markets",
		"list_strategy_kinds",
	}
	slices.Sort(names)
	if !slices.Equal(names, want) {
		t.Fatalf("the tool list is fixed: got %v, want %v", names, want)
	}
	forbidden := regexp.MustCompile(
		`(?i)order|trade|ticket|buy|sell|place|cancel|control|halt|resume|policy|enable|toggle|strategy_create|flatten|reset`,
	)
	for _, name := range names {
		if forbidden.MatchString(name) {
			t.Fatalf("tool %q looks like it changes state", name)
		}
	}
	for _, blocked := range []string{"place_order", "create_ticket", "set_policy", "halt_account", "resume_account"} {
		c.PostJSON("/mcp", rpc("tools/call", map[string]any{"name": blocked, "arguments": map[string]any{}})).
			AssertOK().AssertJSONPath("result.isError", true)
	}
}

func TestMCPReadsMarketsAndMetrics(t *testing.T) {
	app := mcpApp(t, "correct-horse-battery-staple-0123456789")
	c := app.Client(t).WithToken("correct-horse-battery-staple-0123456789")
	call := func(name string, args map[string]any) *vukemtest.Response {
		return c.PostJSON("/mcp", rpc("tools/call", map[string]any{"name": name, "arguments": args})).AssertOK()
	}
	call("list_markets", map[string]any{}).AssertJSONCount("result.content", 1)
	call("list_instruments", map[string]any{}).AssertJSONCount("result.content", 1)
	call("get_backtest_report", map[string]any{"id": "missing"}).AssertJSONPath("result.isError", true)
	call("get_backtest_report", map[string]any{}).AssertJSONPath("result.isError", true)
	c.PostJSON("/mcp", rpc("no/such/method", map[string]any{})).AssertOK().AssertJSONPath("error.code", -32601)
}

func TestSecurity_TR002_MCPThrottlesTokenGuessing(t *testing.T) {
	app := mcpApp(t, "correct-horse-battery-staple-0123456789")
	c := app.Client(t).WithToken("guess")
	var last int
	for range 80 {
		last = c.PostJSON("/mcp", rpc("tools/list", map[string]any{})).Status()
	}
	if last != 429 {
		t.Fatalf("guessing must be throttled, last status %d", last)
	}
}

func TestSecurity_TR002_MCPRejectsAWeakToken(t *testing.T) {
	for _, weak := range []string{"short", "correct-horse-battery"} {
		if err := terminal.ValidateMCPToken(weak); err == nil {
			t.Fatalf("token %q must be refused", weak)
		}
	}
	if err := terminal.ValidateMCPToken(""); err != nil {
		t.Fatalf("an unset token keeps the endpoint off: %v", err)
	}
	if err := terminal.ValidateMCPToken("correct-horse-battery-staple-0123456789"); err != nil {
		t.Fatal(err)
	}
}
