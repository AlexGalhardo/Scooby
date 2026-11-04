package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	d "github.com/shopspring/decimal"

	"github.com/AlexGalhardo/vukem-trading/internal/trading"
)

const (
	binanceBase     = "https://api.binance.com"
	binancePageSize = 1000
	binanceMaxBody  = 4 << 20
	binanceMaxSpan  = 30 * 24 * time.Hour
)

// Binance serves closed 5-minute candles of crypto pairs from Binance's public market-data API. It
// needs no credential. Use is subject to Binance's market-data terms (personal, non-commercial use).
type Binance struct {
	BaseURL  string       // overridden by tests
	Client   *http.Client // optional
	PageSize int          // optional, defaults to 1000
}

// Name implements Provider.
func (Binance) Name() string { return "binance" }

// Markets implements Provider.
func (Binance) Markets() []string { return []string{"CRYPTO"} }

// Requires implements Provider: the klines endpoint is public.
func (Binance) Requires() []string { return nil }

// Bars implements Provider.
func (b Binance) Bars(ctx context.Context, _ Credentials, q Query) (trading.Dataset, error) {
	inst, err := b.transformQuery(q)
	if err != nil {
		return trading.Dataset{}, err
	}
	var candles []trading.Candle
	cursor := q.From.UTC()
	for cursor.Before(q.To) {
		page, err := b.extract(ctx, inst.Symbol, cursor, q.To.UTC())
		if err != nil {
			return trading.Dataset{}, err
		}
		if len(page) == 0 {
			break
		}
		got, err := transformKlines(page)
		if err != nil {
			return trading.Dataset{}, err
		}
		candles = append(candles, got...)
		last := got[len(got)-1].Time
		if !last.After(cursor) && len(candles) > 1 {
			break
		}
		cursor = last.Add(5 * time.Minute)
	}
	// Keep only closed candles: the one still forming has a close time in the future.
	for len(candles) > 0 && candles[len(candles)-1].Time.Add(5*time.Minute).After(q.To.UTC()) {
		candles = candles[:len(candles)-1]
	}
	if len(candles) < 3 {
		return trading.Dataset{}, errors.New("fewer than three closed candles in the period")
	}
	return trading.Dataset{
		CalendarSource: "Crypto: 24/7, UTC days",
		SessionCloses:  map[string]time.Time{},
		ID:             trading.ID(),
		Symbol:         inst.Symbol,
		Source:         "Binance public klines (5m, closed candles)",
		License:        "Binance market-data terms; personal, non-commercial use",
		Hash:           trading.DatasetHash(candles),
		Currency:       inst.Currency,
		Timeframe:      5,
		CreatedAt:      time.Now().UTC(),
		Candles:        candles,
	}, nil
}

func (b Binance) transformQuery(q Query) (trading.Instrument, error) {
	inst, err := trading.FindInstrument(q.Symbol)
	if err != nil || inst.Market != "CRYPTO" {
		return trading.Instrument{}, errors.New("symbol is not a supported crypto pair")
	}
	if !q.From.Before(q.To) || q.To.Sub(q.From) > binanceMaxSpan {
		return trading.Instrument{}, errors.New("period must be between five minutes and 30 days")
	}
	return inst, nil
}

func (b Binance) extract(ctx context.Context, symbol string, from, to time.Time) ([][]json.RawMessage, error) {
	base := b.BaseURL
	if base == "" {
		base = binanceBase
	}
	size := b.PageSize
	if size <= 0 {
		size = binancePageSize
	}
	params := url.Values{
		"symbol":    {symbol},
		"interval":  {"5m"},
		"startTime": {strconv.FormatInt(from.UnixMilli(), 10)},
		"endTime":   {strconv.FormatInt(to.UnixMilli(), 10)},
		"limit":     {strconv.Itoa(size)},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/v3/klines?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	client := b.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("binance request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance returned HTTP %d", resp.StatusCode)
	}
	var rows [][]json.RawMessage
	if err := json.NewDecoder(io.LimitReader(resp.Body, binanceMaxBody)).Decode(&rows); err != nil {
		return nil, errors.New("binance returned an unreadable response")
	}
	return rows, nil
}

// transformKlines maps Binance's array rows [openTime, open, high, low, close, volume, ...] to candles.
func transformKlines(rows [][]json.RawMessage) ([]trading.Candle, error) {
	out := make([]trading.Candle, 0, len(rows))
	for _, row := range rows {
		if len(row) < 6 {
			return nil, errors.New("binance row has too few fields")
		}
		var ms int64
		if err := json.Unmarshal(row[0], &ms); err != nil {
			return nil, errors.New("binance row has an invalid time")
		}
		var vals [5]d.Decimal
		for i := range vals {
			var s string
			if err := json.Unmarshal(row[i+1], &s); err != nil {
				return nil, errors.New("binance row has an invalid number")
			}
			v, err := d.NewFromString(s)
			if err != nil {
				return nil, errors.New("binance row has an invalid number")
			}
			vals[i] = v
		}
		c := trading.Candle{
			Time: time.UnixMilli(ms).UTC(), Open: vals[0], High: vals[1], Low: vals[2], Close: vals[3], Volume: vals[4],
		}
		if err := c.Validate(); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}
