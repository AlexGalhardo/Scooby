package trading

import (
	"math"

	d "github.com/shopspring/decimal"
)

// sessionStart returns the index of the first candle of the last candle's market day.
func sessionStart(cs []Candle, market string) int {
	i := len(cs) - 1
	day := MarketDay(market, cs[i].Time)
	for i > 0 && MarketDay(market, cs[i-1].Time) == day {
		i--
	}
	return i
}

func marketOf(s Strategy) string {
	inst, _ := FindInstrument(s.Symbol)
	return inst.Market
}

// openingRangeSignal buys the first close above the opening range's high and exits on the first
// close below its low. The range is the first s.Fast bars of the session; nothing trades inside it.
func openingRangeSignal(s Strategy, cs []Candle) string {
	n := len(cs)
	start := sessionStart(cs, marketOf(s))
	if n-1-start < s.Fast {
		return "hold"
	}
	high, low := cs[start].High, cs[start].Low
	for _, c := range cs[start+1 : start+s.Fast] {
		high, low = d.Max(high, c.High), d.Min(low, c.Low)
	}
	prev, cur := cs[n-2].Close, cs[n-1].Close
	if cur.GreaterThan(high) && !prev.GreaterThan(high) {
		return "buy"
	}
	if cur.LessThan(low) && !prev.LessThan(low) {
		return "sell"
	}
	return "hold"
}

// vwapSignal buys a close at least s.Threshold percent below the session VWAP and exits at the VWAP.
func vwapSignal(s Strategy, cs []Candle) string {
	n := len(cs)
	start := sessionStart(cs, marketOf(s))
	if n-start < s.Fast {
		return "hold"
	}
	var pv, vol float64
	for _, c := range cs[start:] {
		typical := (c.High.InexactFloat64() + c.Low.InexactFloat64() + c.Close.InexactFloat64()) / 3
		v := c.Volume.InexactFloat64()
		pv += typical * v
		vol += v
	}
	if vol <= 0 {
		return "hold"
	}
	vwap, last := pv/vol, cs[n-1].Close.InexactFloat64()
	switch {
	case last <= vwap*(1-s.Threshold/100):
		return "buy"
	case last >= vwap:
		return "sell"
	}
	return "hold"
}

// directional computes Wilder's +DI, -DI and ADX for the last two bars of cs.
func directional(cs []Candle, period int) (plus, minus, adx, prevPlus, prevMinus float64, ok bool) {
	n := len(cs)
	if period < 2 || n < 2*period+2 {
		return
	}
	var smTR, smPlus, smMinus, adxv float64
	var plusDI, minusDI [2]float64
	dxs := make([]float64, 0, n)
	for i := 1; i < n; i++ {
		h, l, pc := cs[i].High.InexactFloat64(), cs[i].Low.InexactFloat64(), cs[i-1].Close.InexactFloat64()
		ph, pl := cs[i-1].High.InexactFloat64(), cs[i-1].Low.InexactFloat64()
		tr := math.Max(h-l, math.Max(math.Abs(h-pc), math.Abs(l-pc)))
		up, down := h-ph, pl-l
		var dmPlus, dmMinus float64
		if up > down && up > 0 {
			dmPlus = up
		}
		if down > up && down > 0 {
			dmMinus = down
		}
		if i <= period {
			smTR, smPlus, smMinus = smTR+tr, smPlus+dmPlus, smMinus+dmMinus
			if i < period {
				continue
			}
		} else {
			p := float64(period)
			smTR, smPlus, smMinus = smTR-smTR/p+tr, smPlus-smPlus/p+dmPlus, smMinus-smMinus/p+dmMinus
		}
		if smTR <= 0 {
			dxs = append(dxs, 0)
			plusDI[0], plusDI[1], minusDI[0], minusDI[1] = plusDI[1], 0, minusDI[1], 0
			continue
		}
		pd, md := 100*smPlus/smTR, 100*smMinus/smTR
		plusDI[0], plusDI[1], minusDI[0], minusDI[1] = plusDI[1], pd, minusDI[1], md
		dx := 0.0
		if pd+md > 0 {
			dx = 100 * math.Abs(pd-md) / (pd + md)
		}
		dxs = append(dxs, dx)
		switch {
		case len(dxs) == period:
			sum := 0.0
			for _, v := range dxs {
				sum += v
			}
			adxv = sum / float64(period)
		case len(dxs) > period:
			adxv = (adxv*float64(period-1) + dx) / float64(period)
		}
	}
	if len(dxs) < period+1 {
		return
	}
	return plusDI[1], minusDI[1], adxv, plusDI[0], minusDI[0], true
}

// adxSignal follows a +DI/-DI crossover only while the ADX is at least s.Threshold.
func adxSignal(s Strategy, cs []Candle) string {
	plus, minus, adx, prevPlus, prevMinus, ok := directional(cs, s.Fast)
	if !ok {
		return "hold"
	}
	if adx >= s.Threshold && plus > minus && prevPlus <= prevMinus {
		return "buy"
	}
	if plus < minus && prevPlus >= prevMinus {
		return "sell"
	}
	return "hold"
}

// dcaSignal flags a dip: a close at least s.Threshold percent below the mean of the last s.Slow
// closes. The spacing between tranches and the tranche cap are enforced by the risk engine.
func dcaSignal(s Strategy, cs []Candle) string {
	n := len(cs)
	if n < s.Slow {
		return "hold"
	}
	if cs[n-1].Close.InexactFloat64() <= mean(cs[n-s.Slow:])*(1-s.Threshold/100) {
		return "buy"
	}
	return "hold"
}
