package trading

import (
	"errors"
	"math"

	d "github.com/shopspring/decimal"
)

// QuoteInput is the state the Avellaneda-Stoikov market-making model needs.
type QuoteInput struct {
	Mid          d.Decimal // current mid price
	Inventory    d.Decimal // signed position in base units: positive is long
	MaxInventory d.Decimal // quote no more on the side that would grow it past this
	Gamma        float64   // risk aversion, in 1/price units; larger skews quotes harder against inventory
	Sigma        float64   // volatility of the mid price per unit of time, in price units
	Horizon      float64   // remaining time in the same units as Sigma, between 0 and 1 of the session
	Kappa        float64   // order-book liquidity parameter; larger means thinner spreads
	Tick         d.Decimal
}

// Quotes are the two prices to rest. A zero Bid or Ask means that side is not quoted.
type Quotes struct {
	Reservation d.Decimal
	Bid, Ask    d.Decimal
}

// AvellanedaStoikov returns a bid and an ask around the inventory-adjusted reservation price. It is the
// quoting math only. Running it needs an order book and a queue-position model, which this project does
// not have (ADR-0016), so nothing sends these quotes anywhere.
func AvellanedaStoikov(in QuoteInput) (Quotes, error) {
	if !in.Mid.IsPositive() || !in.Tick.IsPositive() || in.Gamma <= 0 || in.Kappa <= 0 || in.Sigma < 0 ||
		in.Horizon < 0 || in.Horizon > 1 || !in.MaxInventory.IsPositive() {
		return Quotes{}, errors.New("invalid quoting inputs")
	}
	mid := in.Mid.InexactFloat64()
	q := in.Inventory.InexactFloat64()
	variance := in.Sigma * in.Sigma
	reservation := mid - q*in.Gamma*variance*in.Horizon
	spread := in.Gamma*variance*in.Horizon + (2/in.Gamma)*math.Log(1+in.Gamma/in.Kappa)
	out := Quotes{Reservation: d.NewFromFloat(reservation)}
	bid := in.Tick.Mul(d.NewFromFloat((reservation - spread/2) / in.Tick.InexactFloat64()).Floor())
	ask := in.Tick.Mul(d.NewFromFloat((reservation + spread/2) / in.Tick.InexactFloat64()).Ceil())
	if in.Inventory.LessThan(in.MaxInventory) && bid.IsPositive() {
		out.Bid = bid
	}
	if in.Inventory.GreaterThan(in.MaxInventory.Neg()) {
		out.Ask = ask
	}
	return out, nil
}
