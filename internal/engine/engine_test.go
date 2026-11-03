package engine_test

import (
	"encoding/json"
	"errors"
	"github.com/AlexGalhardo/Scooby/bootstrap"
	"github.com/AlexGalhardo/Scooby/internal/adapters"
	"github.com/AlexGalhardo/Scooby/internal/engine"
	"github.com/AlexGalhardo/Scooby/internal/trading"
	"github.com/AlexGalhardo/vukem/vukemtest"
	d "github.com/shopspring/decimal"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestTimeoutAfterSendNeverResubmitsAndReconcilesPartial(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New)
	db, _ := app.DB()
	store := trading.Store{DB: db}
	if err := store.Initialize(t.Context()); err != nil {
		t.Fatal(err)
	}
	var order trading.Order
	if err := store.Mutate(t.Context(), "us", "test", "setup", func(a *trading.State) error {
		a.Cash = trading.D("100000")
		a.Initial = a.Cash
		a.DayStart = a.Cash
		a.Halted = false
		o, err := a.Reserve(trading.Strategy{Symbol: "AAPL", StopBPS: 100, TargetBPS: 200}, "buy", "signal", trading.D("100"), time.Now())
		if err == nil {
			order = *o
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	sends := 0
	accountID := "paper-account"
	filled := order.Qty.Div(trading.D("2")).Floor()
	remote := adapters.BrokerOrder{ID: "broker-order", ClientID: order.ID, Symbol: "AAPL", Side: "buy", Qty: order.Qty, Filled: filled, Average: ptr(trading.D("100")), Status: "partially_filled", Updated: time.Now()}
	broker := adapters.Alpaca{Key: "test", Secret: "test", Client: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		if r.Method == "POST" {
			sends++
			return nil, errors.New("timeout after server accepted")
		}
		var body any
		switch r.URL.Path {
		case "/v2/account":
			body = map[string]any{"id": accountID, "cash": "50000", "equity": "100000", "currency": "USD", "status": "ACTIVE"}
		case "/v2/positions":
			body = []map[string]any{{"symbol": "AAPL", "qty": filled.String(), "avg_entry_price": "100", "current_price": "100"}}
		case "/v2/orders":
			body = []adapters.BrokerOrder{remote}
		case "/v2/orders:by_client_order_id":
			body = remote
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		b, _ := json.Marshal(body)
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(b)))}, nil
	})}}
	e := engine.Engine{Store: store, Broker: broker, Log: slog.Default()}
	if err := e.Dispatch(t.Context()); err == nil {
		t.Fatal("timeout must remain unknown")
	}
	var state trading.State
	if err := store.Get(t.Context(), "trading_accounts", "us", &state); err != nil {
		t.Fatal(err)
	}
	if !state.Halted || state.Orders[0].Status != "unknown" {
		t.Fatal("not quarantined")
	}
	_ = e.Dispatch(t.Context())
	if sends != 1 {
		t.Fatal("unknown order resent")
	}
	if err := e.Reconcile(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := e.Reconcile(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := store.Get(t.Context(), "trading_accounts", "us", &state); err != nil {
		t.Fatal(err)
	}
	if len(state.Fills) != 1 || !state.Positions["AAPL"].Qty.Equal(filled) || state.Orders[0].Status != "partially_filled" || !state.Halted {
		t.Fatal("partial/duplicate reconciliation failed")
	}
	accountID = "other-account"
	if err := e.Reconcile(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := store.Get(t.Context(), "trading_accounts", "us", &state); err != nil {
		t.Fatal(err)
	}
	if state.BrokerAccountID != "paper-account" || !strings.Contains(state.Reason, "outra conta") {
		t.Fatal("credential switch mixed broker accounts")
	}
}
func ptr(v d.Decimal) *d.Decimal { return &v }
