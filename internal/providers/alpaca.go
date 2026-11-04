package providers

import (
	"context"

	"github.com/AlexGalhardo/vukem-trading/internal/adapters"
	"github.com/AlexGalhardo/vukem-trading/internal/trading"
)

// AlpacaIEX serves US equity candles from Alpaca's IEX feed. IEX is not the consolidated market.
type AlpacaIEX struct{}

// Name implements Provider; the credential variables are ALPACA_API_KEY and ALPACA_API_SECRET.
func (AlpacaIEX) Name() string { return "alpaca" }

// Markets implements Provider.
func (AlpacaIEX) Markets() []string { return []string{"US"} }

// Requires implements Provider.
func (AlpacaIEX) Requires() []string { return []string{"api_key", "api_secret"} }

// Bars implements Provider.
func (AlpacaIEX) Bars(ctx context.Context, creds Credentials, q Query) (trading.Dataset, error) {
	broker := adapters.Alpaca{Key: creds["api_key"], Secret: creds["api_secret"]}
	return broker.Bars(ctx, q.Symbol, q.From, q.To)
}
