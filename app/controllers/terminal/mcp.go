package terminal

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/vukem-trading/internal/trading"
)

// mcpTool describes one tool the read-only MCP endpoint offers. The list is fixed in code: an agent
// can read market data, reports and operational metrics, and nothing else. There is no tool that
// places, changes or cancels an order, changes a policy, or starts or stops an account.
type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

var noArgs = map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}

func mcpTools() []mcpTool {
	return []mcpTool{
		{"list_markets", "Coverage per market: data mode, what can execute, source and notes.", noArgs},
		{"list_instruments", "The instruments of the watchlist with their market, currency, lot and tick.", noArgs},
		{"list_strategy_kinds", "Strategy kinds and the declared range of each parameter.", noArgs},
		{"list_datasets", "Dataset metadata: symbol, source, license, hash, bars and period. No candles.", noArgs},
		{
			"get_backtest_report",
			"Summary of a finished backtest: metrics, robustness checks and walk-forward windows. No curve or fills.",
			map[string]any{
				"type":                 "object",
				"properties":           map[string]any{"id": map[string]any{"type": "string"}},
				"required":             []string{"id"},
				"additionalProperties": false,
			},
		},
		{
			"get_ops_metrics",
			"Operational metrics of every account: heartbeat, reconciliation, rejections, slippage.",
			noArgs,
		},
	}
}

// MinMCPTokenLength is the shortest MCP_TOKEN the server accepts; a shorter one is a boot error.
const MinMCPTokenLength = 32

// ValidateMCPToken fails closed on a configured but weak token. An empty token keeps the endpoint off.
func ValidateMCPToken(token string) error {
	if token != "" && len(token) < MinMCPTokenLength {
		return errors.New("MCP_TOKEN must be at least 32 characters; generate one with: openssl rand -base64 32")
	}
	return nil
}

type rpcRequest struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

func rpcResult(id json.RawMessage, result any) map[string]any {
	return map[string]any{"jsonrpc": "2.0", "id": id, "result": result}
}

func rpcError(id json.RawMessage, code int, message string) map[string]any {
	return map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": message}}
}

// MCP serves a read-only Model Context Protocol endpoint (JSON-RPC 2.0 over HTTP). It is off unless
// MCP_TOKEN is set, and every call must carry that token as a Bearer credential.
func (h *Controller) MCP(c *web.Ctx) error {
	token := h.App.Source.Get("MCP_TOKEN")
	if token == "" {
		return errs.New(404, "Not found")
	}
	given := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	// Hashing first makes the comparison constant-time regardless of the lengths involved.
	want, got := sha256.Sum256([]byte(token)), sha256.Sum256([]byte(given))
	if !strings.HasPrefix(c.Request().Header.Get("Authorization"), "Bearer ") ||
		subtle.ConstantTimeCompare(want[:], got[:]) != 1 {
		return errs.New(401, "Invalid or missing token")
	}
	var req rpcRequest
	if err := decodeLoose(c, &req); err != nil {
		return c.JSON(200, rpcError(nil, -32700, "Parse error"))
	}
	switch req.Method {
	case "initialize":
		return c.JSON(200, rpcResult(req.ID, map[string]any{
			"protocolVersion": "2025-06-18",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "vukem-trading", "version": trading.BuildVersion},
		}))
	case "notifications/initialized":
		return c.JSON(202, map[string]any{})
	case "ping":
		return c.JSON(200, rpcResult(req.ID, map[string]any{}))
	case "tools/list":
		return c.JSON(200, rpcResult(req.ID, map[string]any{"tools": mcpTools()}))
	case "tools/call":
		var call struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &call); err != nil {
			return c.JSON(200, rpcError(req.ID, -32602, "Invalid params"))
		}
		out, err := h.callTool(c, call.Name, call.Arguments)
		if err != nil {
			message := "Internal error"
			var public *errs.HTTPError
			if errors.As(err, &public) {
				message = public.Text()
			} else {
				h.App.Log.Error("mcp tool failed", "tool", call.Name, "error", err)
			}
			return c.JSON(200, rpcResult(req.ID, map[string]any{
				"isError": true,
				"content": []map[string]any{{"type": "text", "text": message}},
			}))
		}
		text, _ := json.Marshal(out)
		return c.JSON(200, rpcResult(req.ID, map[string]any{
			"content": []map[string]any{{"type": "text", "text": string(text)}},
		}))
	}
	return c.JSON(200, rpcError(req.ID, -32601, "Method not found"))
}

// decodeLoose reads a JSON-RPC request; unlike decode it tolerates the extra fields the protocol defines.
func decodeLoose(c *web.Ctx, out any) error {
	r := json.NewDecoder(io.LimitReader(c.Request().Body, 64<<10))
	return r.Decode(out)
}

func (h *Controller) callTool(c *web.Ctx, name string, args json.RawMessage) (any, error) {
	ctx := c.Context()
	switch name {
	case "list_markets":
		return trading.CoverageMatrix(), nil
	case "list_instruments":
		return trading.Instruments, nil
	case "list_strategy_kinds":
		return trading.Kinds(), nil
	case "list_datasets":
		records, err := h.Store.List(ctx, "datasets", 100)
		if err != nil {
			return nil, err
		}
		out := []map[string]any{}
		for _, r := range records {
			var ds trading.Dataset
			if json.Unmarshal(r.Payload, &ds) != nil {
				continue
			}
			item := map[string]any{
				"id": ds.ID, "symbol": ds.Symbol, "source": ds.Source, "license": ds.License,
				"hash": ds.Hash, "bars": len(ds.Candles), "currency": ds.Currency,
			}
			if len(ds.Candles) > 0 {
				item["first"], item["last"] = ds.Candles[0].Time, ds.Candles[len(ds.Candles)-1].Time
			}
			out = append(out, item)
		}
		return out, nil
	case "get_backtest_report":
		var in struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(args, &in); err != nil || in.ID == "" {
			return nil, errs.New(422, "id is required")
		}
		var job trading.Job
		if err := h.Store.Get(ctx, "backtests", in.ID, &job); err != nil {
			return nil, errs.New(404, "Backtest not found")
		}
		summary := map[string]any{
			"id":        job.ID,
			"strategy":  job.Strategy,
			"partition": job.Partition,
			"error":     job.Error,
		}
		if job.Result != nil {
			summary["metrics"] = job.Result.Metrics
			summary["robustness"] = job.Result.Robustness
			summary["windows"] = job.Result.Windows
			summary["warnings"] = job.Result.Warnings
		}
		return summary, nil
	case "get_ops_metrics":
		records, err := h.Store.List(ctx, "trading_accounts", 10)
		if err != nil {
			return nil, err
		}
		now := time.Now().UTC()
		out := []trading.OpsMetrics{}
		for _, r := range records {
			var a trading.State
			if json.Unmarshal(r.Payload, &a) == nil {
				out = append(out, trading.OperationalMetrics(a, now))
			}
		}
		return out, nil
	}
	return nil, errs.New(404, "Unknown tool")
}
