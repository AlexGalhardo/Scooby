package adapters

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	d "github.com/shopspring/decimal"
)

const testSecret = "unit-test-secret"

// signedServer verifies the signature the way Binance does and records what it saw.
func signedServer(t *testing.T, requests *atomic.Int32, last *url.Values, reply string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Header.Get("X-MBX-APIKEY") != "unit-test-key" {
			http.Error(w, "no key", http.StatusUnauthorized)
			return
		}
		query := r.URL.Query()
		signature := query.Get("signature")
		query.Del("signature")
		mac := hmac.New(sha256.New, []byte(testSecret))
		_, _ = mac.Write([]byte(query.Encode()))
		if signature != hex.EncodeToString(mac.Sum(nil)) {
			http.Error(w, "bad signature", http.StatusBadRequest)
			return
		}
		*last = query
		_, _ = w.Write([]byte(reply))
	}))
}

func testnet(base string) BinanceTestnet {
	fixed := time.Date(2026, 11, 4, 15, 0, 0, 0, time.UTC)
	return BinanceTestnet{
		Key:     "unit-test-key",
		Secret:  testSecret,
		BaseURL: base,
		Now:     func() time.Time { return fixed },
	}
}

func TestTestnetSignsAndSendsAnIdempotentLimitOrder(t *testing.T) {
	var requests atomic.Int32
	var seen url.Values
	srv := signedServer(
		t,
		&requests,
		&seen,
		`{"orderId":7,"clientOrderId":"vt-1","status":"NEW","executedQty":"0.00000000","cummulativeQuoteQty":"0.00000000"}`,
	)
	defer srv.Close()
	ack, err := testnet(
		srv.URL,
	).Submit(t.Context(), "BTCUSDT", "BUY", d.RequireFromString("0.01"), d.RequireFromString("60000"), "vt-1")
	if err != nil {
		t.Fatal(err)
	}
	if ack.OrderID != 7 || ack.ClientOrderID != "vt-1" || ack.Status != "NEW" || !ack.Filled.IsZero() {
		t.Fatalf("unexpected ack: %+v", ack)
	}
	for key, want := range map[string]string{
		"symbol": "BTCUSDT", "side": "BUY", "type": "LIMIT", "timeInForce": "GTC",
		"quantity": "0.01", "price": "60000", "newClientOrderId": "vt-1",
	} {
		if seen.Get(key) != want {
			t.Fatalf("%s = %q, want %q", key, seen.Get(key), want)
		}
	}
	if seen.Get("timestamp") != "1793804400000" {
		t.Fatalf("the injected clock must drive the timestamp, got %s", seen.Get("timestamp"))
	}
}

func TestTestnetLookupAndCancelUseTheClientOrderID(t *testing.T) {
	var requests atomic.Int32
	var seen url.Values
	srv := signedServer(
		t,
		&requests,
		&seen,
		`{"orderId":7,"origClientOrderId":"vt-1","status":"PARTIALLY_FILLED","executedQty":"0.004","cummulativeQuoteQty":"240.0"}`,
	)
	defer srv.Close()
	b := testnet(srv.URL)
	got, err := b.Lookup(t.Context(), "BTCUSDT", "vt-1")
	if err != nil || got.Status != "PARTIALLY_FILLED" || !got.Filled.Equal(d.RequireFromString("0.004")) ||
		got.ClientOrderID != "vt-1" {
		t.Fatalf("lookup: %+v %v", got, err)
	}
	if seen.Get("origClientOrderId") != "vt-1" {
		t.Fatal("a lookup is keyed by the client order id")
	}
	if _, err := b.Cancel(t.Context(), "BTCUSDT", "vt-1"); err != nil {
		t.Fatal(err)
	}
}

func TestTestnetRefusesEveryOtherHostBeforeAnyRequest(t *testing.T) {
	for _, base := range []string{"https://api.binance.com", "https://testnet.binance.vision.evil.example", "http://10.0.0.5:9000"} {
		b := testnet(base)
		if _, err := b.Balances(
			t.Context(),
		); err == nil ||
			!strings.Contains(err.Error(), "only the Binance spot testnet") {
			t.Fatalf("%s must be refused, got %v", base, err)
		}
	}
	if _, err := (BinanceTestnet{BaseURL: "https://testnet.binance.vision"}).Balances(t.Context()); err == nil ||
		!strings.Contains(err.Error(), "credentials") {
		t.Fatalf("missing credentials must fail before any call, got %v", err)
	}
}

func TestTestnetErrorsNeverEchoTheSignedRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, r.URL.RawQuery, http.StatusBadRequest)
	}))
	defer srv.Close()
	_, err := testnet(srv.URL).Lookup(t.Context(), "BTCUSDT", "vt-1")
	if err == nil || strings.Contains(err.Error(), "signature") || strings.Contains(err.Error(), "vt-1") ||
		strings.Contains(err.Error(), testSecret) {
		t.Fatalf("an error must not repeat the request or any secret, got %v", err)
	}
}

func TestTestnetBalancesKeepOnlyFundedAssets(t *testing.T) {
	var requests atomic.Int32
	var seen url.Values
	srv := signedServer(t, &requests, &seen,
		`{"canTrade":true,"balances":[{"asset":"USDT","free":"10000.5"},{"asset":"BTC","free":"0.00000000"}]}`)
	defer srv.Close()
	got, err := testnet(srv.URL).Balances(t.Context())
	if err != nil || len(got) != 1 || !got["USDT"].Equal(d.RequireFromString("10000.5")) {
		t.Fatalf("balances: %v %v", got, err)
	}
	if _, err := testnet(
		srv.URL,
	).Submit(t.Context(), "BTCUSDT", "HOLD", d.NewFromInt(1), d.NewFromInt(1), "x"); err == nil {
		t.Fatal("an invalid side must be rejected before any request")
	}
}
