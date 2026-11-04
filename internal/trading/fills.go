package trading

import (
	"strconv"

	d "github.com/shopspring/decimal"
)

// ParticipationRate is the largest share of a bar's volume one order may take.
var ParticipationRate = D("0.01")

// restingFill decides whether an open order fills on candle c and at which price (ADR 0016).
//
// A marketable order fills at the open plus slippage, never beyond the bar's range. A limit the open
// does not reach keeps resting and fills at the limit only when the bar trades strictly through it:
// merely touching the limit gives no queue priority, so it does not fill.
func restingFill(p Policy, o Order, c Candle) (d.Decimal, bool) {
	if o.Side == "buy" {
		price := d.Min(c.Open.Mul(D("1").Add(Fraction(p.SlippageBPS))), c.High)
		if !price.GreaterThan(o.Limit) {
			return price, true
		}
		if c.Low.LessThan(o.Limit) {
			return o.Limit, true
		}
		return Zero, false
	}
	price := d.Max(c.Open.Mul(D("1").Sub(Fraction(p.SlippageBPS))), c.Low)
	if !price.LessThan(o.Limit) {
		return price, true
	}
	if c.High.GreaterThan(o.Limit) {
		return o.Limit, true
	}
	return Zero, false
}

// Simulate offers candle c to an open order. It fills up to the participation cap and leaves the
// rest open for the next bars of the same session; the caller expires orders at the session end.
func Simulate(a *State, o *Order, c Candle) error {
	if o == nil || !o.Open() {
		return nil
	}
	price, ok := restingFill(a.Policy, *o, c)
	if !ok {
		return nil
	}
	instrument, err := FindInstrument(o.Symbol)
	if err != nil {
		return err
	}
	room := c.Volume.Mul(ParticipationRate).Div(instrument.Step).Floor().Mul(instrument.Step)
	qty := d.Min(o.Qty.Sub(o.Filled), room)
	if !qty.IsPositive() {
		return nil
	}
	return a.ApplyFill(Fill{
		ID:      o.ID + "/" + strconv.FormatInt(c.Time.Unix(), 10),
		OrderID: o.ID,
		Symbol:  o.Symbol,
		Side:    o.Side,
		Qty:     qty,
		Price:   price,
		Fees:    qty.Mul(price).Mul(Fraction(a.Policy.CostBPS)),
		Time:    c.Time,
	})
}
