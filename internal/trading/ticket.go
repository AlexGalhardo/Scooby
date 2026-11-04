package trading

import (
	"errors"
	"time"

	d "github.com/shopspring/decimal"
)

// Ticket turns an order template into a paper order. It goes through the same risk engine as a
// strategy entry, then trims the size to the template's share of the available cash: the order is
// never larger than either limit.
func (a *State) Ticket(t OrderTemplate, symbol string, price d.Decimal, key string, at time.Time) (*Order, error) {
	inst, err := FindInstrument(symbol)
	if err != nil {
		return nil, err
	}
	s := Strategy{
		Name: "ticket", Kind: "manual", Symbol: symbol, AccountID: a.ID, Fast: 2, Slow: 3,
		StopBPS: t.StopBPS, TargetBPS: t.TargetBPS,
	}
	if t.Side == "close" {
		o, err := a.Reserve(s, "sell", key, price, at)
		if err != nil {
			return nil, err
		}
		o.Close = CloseManual
		return o, nil
	}
	o, err := a.Reserve(s, "buy", key, price, at)
	if err != nil {
		return nil, err
	}
	budget := a.Cash.Mul(Fraction(t.CashPct * 100)) // a percentage is 100 basis points per point
	perUnit := o.Limit.Mul(D("1").Add(Fraction(a.Policy.CostBPS)))
	room := budget.Div(perUnit).Div(inst.Step).Floor().Mul(inst.Step)
	if room.LessThan(o.Qty) {
		o.Qty = room
		o.Reserved = o.Qty.Mul(perUnit)
	}
	if !o.Qty.IsPositive() {
		a.Orders = a.Orders[:len(a.Orders)-1]
		return nil, errors.New("ticket size below the minimum lot")
	}
	return o, nil
}
