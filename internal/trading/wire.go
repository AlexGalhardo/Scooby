package trading

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	d "github.com/shopspring/decimal"
)

// wireTypes are the Go structs the web terminal receives, in output order. Adding a struct to the API
// means adding it here; the test that compares the generated file fails until it is regenerated.
var wireTypes = []any{
	Candle{},
	Instrument{},
	Strategy{},
	Policy{},
	Order{},
	Position{},
	Fill{},
	Point{},
	Metrics{},
	Window{},
	MonteCarlo{},
	Robustness{},
	Result{},
	ParamSpec{},
	KindSpec{},
	Coverage{},
	PriceAlert{},
	OrderTemplate{},
	Preferences{},
}

// WireTypeScript renders the TypeScript declarations of the wire structs, following their JSON tags.
// Decimals and times travel as strings; a nil slice travels as null; omitempty makes a field optional.
func WireTypeScript() string {
	known := map[string]bool{}
	for _, v := range wireTypes {
		known[reflect.TypeOf(v).Name()] = true
	}
	var b strings.Builder
	b.WriteString(
		"// Code generated from internal/trading by `go test ./internal/trading -run Wire -update`. DO NOT EDIT.\n\n",
	)
	b.WriteString("export type Decimal = string;\n")
	for _, v := range wireTypes {
		t := reflect.TypeOf(v)
		fmt.Fprintf(&b, "\nexport interface %s {\n", t.Name())
		for f := range t.Fields() {
			name, opts, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "-" || name == "" {
				continue
			}
			ts, nullable := tsType(f.Type, known)
			optional := strings.Contains(opts, "omitempty")
			switch {
			case optional:
				fmt.Fprintf(&b, "  %s?: %s;\n", name, ts)
			case nullable:
				fmt.Fprintf(&b, "  %s: %s | null;\n", name, ts)
			default:
				fmt.Fprintf(&b, "  %s: %s;\n", name, ts)
			}
		}
		b.WriteString("}\n")
	}
	return b.String()
}

// tsType maps a Go type to its TypeScript spelling and whether its JSON form can be null.
func tsType(t reflect.Type, known map[string]bool) (string, bool) {
	switch t {
	case reflect.TypeFor[d.Decimal]():
		return "Decimal", false
	case reflect.TypeFor[time.Time]():
		return "string", false
	}
	switch t.Kind() {
	case reflect.String:
		return "string", false
	case reflect.Bool:
		return "boolean", false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return "number", false
	case reflect.Slice:
		inner, _ := tsType(t.Elem(), known)
		if strings.ContainsAny(inner, " |") {
			inner = "(" + inner + ")"
		}
		return inner + "[]", true
	case reflect.Map:
		inner, _ := tsType(t.Elem(), known)
		return "Record<string, " + inner + ">", true
	case reflect.Pointer:
		inner, _ := tsType(t.Elem(), known)
		return inner, true
	case reflect.Struct:
		if !known[t.Name()] {
			panic("wire struct " + t.Name() + " is not in wireTypes")
		}
		return t.Name(), false
	default:
		panic("wire: unsupported type " + t.String())
	}
}
