package trading

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	d "github.com/shopspring/decimal"
	"io"
	"strings"
	"time"
)

// ImportCSV is strict and atomic: a rejected row rejects the whole dataset.
func ImportCSV(input io.Reader, symbol, source, license string) (Dataset, error) {
	i, err := FindInstrument(symbol)
	if err != nil {
		return Dataset{}, err
	}
	if strings.TrimSpace(source) == "" || strings.TrimSpace(license) == "" {
		return Dataset{}, errors.New("source and license are required")
	}
	if i.Market != "B3" {
		return Dataset{}, errors.New("CSV is for B3 replay; use Alpaca history for US")
	}
	raw, err := io.ReadAll(io.LimitReader(input, 10*1024*1024+1))
	if err != nil {
		return Dataset{}, err
	}
	if len(raw) > 10*1024*1024 {
		return Dataset{}, errors.New("CSV exceeds 10 MiB")
	}
	hash := sha256.Sum256(raw)
	ds := Dataset{CalendarSource: B3CalendarVersion, SessionCloses: map[string]time.Time{}, ID: ID(), Symbol: symbol, Source: source, License: license, Hash: hex.EncodeToString(hash[:]), Currency: i.Currency, Timeframe: 5, CreatedAt: time.Now().UTC(), Candles: []Candle{}}
	r := csv.NewReader(strings.NewReader(string(raw)))
	header, err := r.Read()
	if err != nil || strings.Join(header, ",") != "time,open,high,low,close,volume" {
		return ds, errors.New("invalid header; expected: time,open,high,low,close,volume")
	}
	for line := 2; ; line++ {
		row, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil || len(row) != 6 {
			return ds, fmt.Errorf("invalid row %d", line)
		}
		t, e := time.Parse(time.RFC3339, row[0])
		if e != nil {
			return ds, fmt.Errorf("row %d requires RFC3339 timestamp with offset", line)
		}
		vals := make([]d.Decimal, 5)
		for j := range vals {
			if len(row[j+1]) > 32 {
				return ds, fmt.Errorf("decimal exceeds limit on row %d", line)
			}
			vals[j], e = d.NewFromString(row[j+1])
			if e != nil || vals[j].Exponent() < -8 || vals[j].Exponent() > 12 || vals[j].Abs().GreaterThan(D("1000000000000")) {
				return ds, fmt.Errorf("invalid decimal on row %d", line)
			}
		}
		c := Candle{t.UTC(), vals[0], vals[1], vals[2], vals[3], vals[4]}
		if e = c.Validate(); e != nil {
			return ds, fmt.Errorf("row %d: %w", line, e)
		}
		if !Session(i.Market, c.Time) || c.Time.Second() != 0 || c.Time.Minute()%5 != 0 {
			return ds, fmt.Errorf("row %d outside the 5-minute session grid (supported B3 calendar: 2026)", line)
		}
		if len(ds.Candles) > 0 && !c.Time.After(ds.Candles[len(ds.Candles)-1].Time) {
			return ds, fmt.Errorf("row %d duplicated or out of order", line)
		}
		_, end, _ := SessionBounds(i.Market, c.Time)
		ds.SessionCloses[MarketDay(i.Market, c.Time)] = end
		ds.Candles = append(ds.Candles, c)
		if len(ds.Candles) > 100000 {
			return ds, errors.New("maximum of 100000 candles")
		}
	}
	if len(ds.Candles) < 3 {
		return ds, errors.New("minimum of three candles")
	}
	return ds, nil
}
func DatasetProblems(ds Dataset) []string {
	issues := []string{}
	for j := 1; j < len(ds.Candles); j++ {
		prev, c := ds.Candles[j-1], ds.Candles[j]
		if c.Time.Format("2006-01-02") == prev.Time.Format("2006-01-02") && c.Time.Sub(prev.Time) > 5*time.Minute {
			issues = append(issues, "Gap at "+c.Time.Format(time.RFC3339))
			if len(issues) >= 10 {
				break
			}
		}
	}
	return issues
}
func DatasetHash(cs []Candle) string {
	h := sha256.New()
	for _, c := range cs {
		fmt.Fprint(h, c.Time.Format(time.RFC3339), c.Open, c.High, c.Low, c.Close, c.Volume)
	}
	return hex.EncodeToString(h.Sum(nil))
}
func SampleCSV() string {
	return "time,open,high,low,close,volume\n2026-10-26T10:00:00-03:00,30.00,30.20,29.90,30.10,100000\n2026-10-26T10:05:00-03:00,30.10,30.30,30.00,30.20,100000\n2026-10-26T10:10:00-03:00,30.20,30.25,30.00,30.05,100000\n"
}
