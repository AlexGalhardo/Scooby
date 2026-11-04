package adapters

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem-trading/internal/trading"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestPaperEndpointAndUnknownOutcome(t *testing.T) {
	calls := 0
	a := Alpaca{
		Key:    "test",
		Secret: "test",
		Client: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
			calls++
			if r.URL.Host != "paper-api.alpaca.markets" {
				t.Fatal("non paper endpoint", r.URL)
			}
			if r.Header.Get("APCA-API-KEY-ID") != "test" {
				t.Fatal("missing credentials")
			}
			if r.Method == http.MethodPost {
				raw, _ := io.ReadAll(r.Body)
				if !strings.Contains(string(raw), "client_order_id") {
					t.Fatal("missing client ID")
				}
			}
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(strings.NewReader("private broker error")),
				Header:     http.Header{},
			}, nil
		})},
	}
	_, err := a.Submit(
		context.Background(),
		trading.Order{ID: "same-client-id", Symbol: "AAPL", Qty: trading.D("1"), Limit: trading.D("10"), Side: "buy"},
	)
	if err == nil || calls != 1 || strings.Contains(err.Error(), "private broker error") {
		t.Fatal("unsafe retry or response disclosure", err)
	}
}

func TestMissingCredentialsNeverCallsNetwork(t *testing.T) {
	a := Alpaca{Client: &http.Client{Transport: transport(func(*http.Request) (*http.Response, error) {
		t.Fatal("network called without credentials")
		return nil, nil
	})}}
	if _, err := a.Account(context.Background()); err == nil {
		t.Fatal("missing config accepted")
	}
}
