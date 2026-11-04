package trading

import (
	"errors"
	"fmt"
	"math"
)

// ParamSpec declares one strategy parameter as data. The validator, the API and the web form all read
// this single description, so a range is written once.
type ParamSpec struct {
	Name    string  `json:"name"` // the Strategy JSON field the value lives in
	Label   string  `json:"label"`
	Type    string  `json:"type"` // "int" or "float"
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Step    float64 `json:"step"`
	Default float64 `json:"default"`
}

// KindSpec describes a strategy kind and the parameters it uses.
type KindSpec struct {
	Kind        string      `json:"kind"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Params      []ParamSpec `json:"params"`
	// Scales marks kinds that may add to an open position, within a capped number of tranches.
	Scales bool `json:"scales"`
}

func stopTarget() []ParamSpec {
	return []ParamSpec{
		{"stop_bps", "Stop distance (bps)", "int", MinStopBPS, 2000, 10, 100},
		{"target_bps", "Target distance (bps)", "int", 10, 5000, 10, 200},
	}
}

// Kinds lists every strategy kind in display order.
func Kinds() []KindSpec {
	return append(baseKinds(), extraKinds()...)
}

func baseKinds() []KindSpec {
	return []KindSpec{
		{
			Kind:        "ma_cross",
			Label:       "Moving-average crossover",
			Description: "Buys when the fast average crosses above the slow one, sells on the opposite cross.",
			Params: append([]ParamSpec{
				{"fast", "Fast window (bars)", "int", 2, 499, 1, 5},
				{"slow", "Slow window (bars)", "int", 3, 500, 1, 20},
				{"threshold", "Threshold (unused by the crossover)", "float", 0.1, 10, 0.1, 2},
			}, stopTarget()...),
		},
		{
			Kind:        "mean_reversion",
			Label:       "Mean reversion (Z-score)",
			Description: "Buys when the close is far below its rolling mean, exits when it returns to the mean.",
			Params: append([]ParamSpec{
				{"fast", "Minimum window (bars)", "int", 2, 499, 1, 5},
				{"slow", "Lookback window (bars)", "int", 3, 500, 1, 20},
				{"threshold", "Entry Z-score", "float", 0.1, 10, 0.1, 2},
			}, stopTarget()...),
		},
		{
			Kind:        "breakout",
			Label:       "Range breakout",
			Description: "Buys a close above the previous range high, sells a close below the range low.",
			Params: append([]ParamSpec{
				{"fast", "Minimum window (bars)", "int", 2, 499, 1, 5},
				{"slow", "Range lookback (bars)", "int", 3, 500, 1, 20},
				{"threshold", "Threshold (unused by the breakout)", "float", 0.1, 10, 0.1, 2},
			}, stopTarget()...),
		},
	}
}

func extraKinds() []KindSpec {
	return []KindSpec{
		{
			Kind:        "opening_range",
			Label:       "Opening-range breakout",
			Description: "Buys the first close above the high of the session's opening range, exits on a close below its low.",
			Params: append([]ParamSpec{
				{"fast", "Opening range (bars)", "int", 2, 60, 1, 6},
				{"slow", "Warm-up (bars)", "int", 3, 500, 1, 12},
			}, stopTarget()...),
		},
		{
			Kind:        "vwap_reversion",
			Label:       "VWAP reversion",
			Description: "Buys a close well below the session VWAP and exits when price returns to it.",
			Params: append([]ParamSpec{
				{"fast", "Minimum bars in the session", "int", 2, 60, 1, 6},
				{"slow", "Warm-up (bars)", "int", 3, 500, 1, 12},
				{"threshold", "Entry distance below VWAP (%)", "float", 0.1, 10, 0.1, 0.5},
			}, stopTarget()...),
		},
		{
			Kind:        "adx_trend",
			Label:       "ADX-filtered trend",
			Description: "Follows a directional-index crossover only while the ADX shows a trend strong enough to trade.",
			Params: append([]ParamSpec{
				{"fast", "ADX period (bars)", "int", 2, 60, 1, 14},
				{"slow", "Warm-up (bars)", "int", 3, 500, 1, 40},
				{"threshold", "Minimum ADX", "float", 10, 50, 1, 25},
			}, stopTarget()...),
		},
		{
			Kind:        "dca_grid",
			Label:       "DCA grid with capped inventory",
			Description: "Buys dips in tranches spaced by a fixed percentage, up to a tranche cap; stop and target apply to the average cost.",
			Scales:      true,
			Params: append([]ParamSpec{
				{"fast", "Maximum tranches", "int", 2, 20, 1, 4},
				{"slow", "Dip lookback (bars)", "int", 3, 500, 1, 20},
				{"threshold", "Dip and tranche spacing (%)", "float", 0.1, 10, 0.1, 1},
			}, []ParamSpec{
				{"stop_bps", "Stop below average cost (bps)", "int", MinStopBPS, 2000, 10, 500},
				{"target_bps", "Target above average cost (bps)", "int", 10, 5000, 10, 300},
			}...),
		},
	}
}

// FindKind returns the spec of a strategy kind.
func FindKind(kind string) (KindSpec, bool) {
	for _, k := range Kinds() {
		if k.Kind == kind {
			return k, true
		}
	}
	return KindSpec{}, false
}

// value reads the parameter a ParamSpec names from the strategy.
func (s Strategy) value(name string) (float64, bool) {
	switch name {
	case "fast":
		return float64(s.Fast), true
	case "slow":
		return float64(s.Slow), true
	case "threshold":
		return s.Threshold, true
	case "stop_bps":
		return float64(s.StopBPS), true
	case "target_bps":
		return float64(s.TargetBPS), true
	}
	return 0, false
}

// validateParams checks every declared parameter of the strategy's kind.
func (s Strategy) validateParams() error {
	spec, ok := FindKind(s.Kind)
	if !ok {
		return errors.New("invalid strategy parameters")
	}
	for _, p := range spec.Params {
		v, known := s.value(p.Name)
		if !known || math.IsNaN(v) || math.IsInf(v, 0) || v < p.Min || v > p.Max {
			return fmt.Errorf("invalid strategy parameters: %s", p.Name)
		}
		if p.Type == "int" && v != math.Trunc(v) {
			return fmt.Errorf("invalid strategy parameters: %s", p.Name)
		}
	}
	if s.Slow <= s.Fast {
		return errors.New("invalid strategy parameters: slow must exceed fast")
	}
	return nil
}

// with returns the strategy with one declared parameter set to v.
func (s Strategy) with(name string, v float64) Strategy {
	switch name {
	case "fast":
		s.Fast = int(v)
	case "slow":
		s.Slow = int(v)
	case "threshold":
		s.Threshold = v
	case "stop_bps":
		s.StopBPS = int(v)
	case "target_bps":
		s.TargetBPS = int(v)
	}
	return s
}

// Defaults returns the strategy with every parameter of its kind set to the declared default.
func (s Strategy) Defaults() Strategy {
	spec, ok := FindKind(s.Kind)
	if !ok {
		return s
	}
	for _, p := range spec.Params {
		s = s.with(p.Name, p.Default)
	}
	return s
}
