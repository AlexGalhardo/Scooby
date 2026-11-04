// Package providers defines how market data enters the terminal: one interface per source, a
// registry, and credentials that are checked before any network call.
package providers

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem-trading/internal/trading"
)

// ErrMissingCredential is returned, before any request is made, when a provider needs a credential
// that is not configured.
var ErrMissingCredential = errors.New("missing credential")

// Credentials maps a credential name (for example "api_key") to its value.
type Credentials map[string]string

// Query asks a provider for the closed 5-minute candles of one symbol between From and To.
type Query struct {
	Symbol   string
	From, To time.Time
}

// Provider is one market-data source. Bars follows three stages: normalize the query, extract the raw
// data from the source, and transform it into the standard Dataset.
type Provider interface {
	// Name is the registry key and the prefix of its credential environment variables.
	Name() string
	// Markets lists the markets the provider can serve.
	Markets() []string
	// Requires lists the credential names the provider needs; empty means public data.
	Requires() []string
	Bars(ctx context.Context, creds Credentials, q Query) (trading.Dataset, error)
}

// Registry holds the configured providers and resolves their credentials from the environment.
type Registry struct {
	providers map[string]Provider
	// Env reads one environment variable; it is injected so tests need no real environment.
	Env func(string) string
}

// NewRegistry builds a registry with the given providers.
func NewRegistry(env func(string) string, ps ...Provider) Registry {
	r := Registry{providers: map[string]Provider{}, Env: env}
	for _, p := range ps {
		r.providers[p.Name()] = p
	}
	return r
}

// For returns the provider that serves a market, preferring the first registered match by name.
func (r Registry) For(market string) (Provider, bool) {
	for _, name := range r.Names() {
		if p := r.providers[name]; slices.Contains(p.Markets(), market) {
			return p, true
		}
	}
	return nil, false
}

// Names lists the registered providers in a stable order.
func (r Registry) Names() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Fetch resolves the provider's credentials and, only if all are present, asks it for the data.
func (r Registry) Fetch(ctx context.Context, name string, q Query) (trading.Dataset, error) {
	p, ok := r.providers[name]
	if !ok {
		return trading.Dataset{}, fmt.Errorf("unknown provider %q", name)
	}
	creds := Credentials{}
	for _, c := range p.Requires() {
		key := strings.ToUpper(name + "_" + c)
		v := ""
		if r.Env != nil {
			v = r.Env(key)
		}
		if v == "" {
			return trading.Dataset{}, fmt.Errorf("%w: set %s", ErrMissingCredential, key)
		}
		creds[c] = v
	}
	return p.Bars(ctx, creds, q)
}
