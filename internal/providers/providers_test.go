package providers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem-trading/internal/trading"
)

type keyed struct{ called *atomic.Int32 }

func (keyed) Name() string       { return "keyed" }
func (keyed) Markets() []string  { return []string{"US"} }
func (keyed) Requires() []string { return []string{"api_key"} }

func (k keyed) Bars(_ context.Context, creds Credentials, _ Query) (trading.Dataset, error) {
	k.called.Add(1)
	if creds["api_key"] != "secret" {
		return trading.Dataset{}, errors.New("credentials not passed through")
	}
	return trading.Dataset{Symbol: "AAPL"}, nil
}

func TestMissingCredentialFailsBeforeAnyCall(t *testing.T) {
	var called atomic.Int32
	reg := NewRegistry(func(string) string { return "" }, keyed{&called})
	_, err := reg.Fetch(t.Context(), "keyed", Query{Symbol: "AAPL"})
	if !errors.Is(err, ErrMissingCredential) {
		t.Fatalf("want ErrMissingCredential, got %v", err)
	}
	if called.Load() != 0 {
		t.Fatal("the provider must not run when a credential is missing")
	}
	reg = NewRegistry(func(k string) string {
		if k == "KEYED_API_KEY" {
			return "secret"
		}
		return ""
	}, keyed{&called})
	if _, err := reg.Fetch(t.Context(), "keyed", Query{Symbol: "AAPL"}); err != nil || called.Load() != 1 {
		t.Fatalf("with the credential set the provider runs once: err=%v calls=%d", err, called.Load())
	}
	if _, err := reg.Fetch(t.Context(), "nope", Query{}); err == nil {
		t.Fatal("an unknown provider must be rejected")
	}
}

func TestRegistryResolvesAProviderByMarket(t *testing.T) {
	reg := NewRegistry(nil, Binance{}, AlpacaIEX{})
	if p, ok := reg.For("CRYPTO"); !ok || p.Name() != "binance" {
		t.Fatalf("crypto must resolve to binance, got %v %v", p, ok)
	}
	if p, ok := reg.For("US"); !ok || p.Name() != "alpaca" {
		t.Fatalf("US must resolve to alpaca, got %v %v", p, ok)
	}
	if _, ok := reg.For("B3"); ok {
		t.Fatal("B3 has no network provider; it is imported from CSV")
	}
}

// fakeBinance serves n 5-minute klines starting at start, honoring startTime, endTime and limit.
func fakeBinance(t *testing.T, start time.Time, n int, requests *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		from, _ := strconv.ParseInt(r.URL.Query().Get("startTime"), 10, 64)
		to, _ := strconv.ParseInt(r.URL.Query().Get("endTime"), 10, 64)
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		var rows [][]any
		for i := range n {
			at := start.Add(time.Duration(i) * 5 * time.Minute).UnixMilli()
			if at < from || at > to || len(rows) >= limit {
				continue
			}
			p := 100 + float64(i)
			rows = append(rows, []any{
				at, strconv.FormatFloat(p, 'f', 2, 64), strconv.FormatFloat(p+1, 'f', 2, 64),
				strconv.FormatFloat(p-1, 'f', 2, 64), strconv.FormatFloat(p+0.5, 'f', 2, 64), "12.5", at + 299999,
			})
		}
		_ = json.NewEncoder(w).Encode(rows)
	}))
}

func TestBinanceParsesPagesAndKeepsOnlyClosedCandles(t *testing.T) {
	var requests atomic.Int32
	start := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	srv := fakeBinance(t, start, 12, &requests)
	defer srv.Close()
	b := Binance{BaseURL: srv.URL, PageSize: 5}
	// The last candle (index 11) closes after "to", so it is still forming and must be dropped.
	to := start.Add(11*5*time.Minute + 2*time.Minute)
	ds, err := b.Bars(t.Context(), nil, Query{Symbol: "BTCUSDT", From: start, To: to})
	if err != nil {
		t.Fatal(err)
	}
	if len(ds.Candles) != 11 {
		t.Fatalf("want 11 closed candles, got %d", len(ds.Candles))
	}
	if requests.Load() < 3 {
		t.Fatalf("a page size of 5 must paginate, got %d requests", requests.Load())
	}
	if ds.Symbol != "BTCUSDT" || ds.Currency != "USDT" || ds.Timeframe != 5 || ds.Hash == "" || ds.Source == "" ||
		ds.License == "" {
		t.Fatalf("the dataset must carry provenance: %+v", ds)
	}
	for i := 1; i < len(ds.Candles); i++ {
		if !ds.Candles[i].Time.After(ds.Candles[i-1].Time) {
			t.Fatal("candles must be strictly ordered")
		}
	}
}

func TestBinanceRejectsBadQueriesWithoutCallingOut(t *testing.T) {
	var requests atomic.Int32
	srv := fakeBinance(t, time.Now(), 3, &requests)
	defer srv.Close()
	b := Binance{BaseURL: srv.URL}
	now := time.Now()
	for name, q := range map[string]Query{
		"not a crypto pair": {Symbol: "AAPL", From: now.Add(-time.Hour), To: now},
		"reversed period":   {Symbol: "BTCUSDT", From: now, To: now.Add(-time.Hour)},
		"too long":          {Symbol: "BTCUSDT", From: now.Add(-40 * 24 * time.Hour), To: now},
	} {
		if _, err := b.Bars(t.Context(), nil, q); err == nil {
			t.Fatalf("%s must be rejected", name)
		}
	}
	if requests.Load() != 0 {
		t.Fatal("a rejected query must not reach the network")
	}
}

func TestEveryProviderIsDocumentedWithItsTerms(t *testing.T) {
	doc, err := os.ReadFile("../../docs/sources.md")
	if err != nil {
		t.Fatal(err)
	}
	reg := NewRegistry(nil, Binance{}, AlpacaIEX{})
	for _, name := range reg.Names() {
		if !strings.Contains(string(doc), "(`"+name+"`)") {
			t.Fatalf("docs/sources.md has no row for provider %q", name)
		}
	}
}
