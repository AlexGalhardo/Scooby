package accounts

import "testing"

func TestValidateAppURL(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name       string
		production bool
		url        string
		ok         bool
	}{
		{"production canonical host", true, "https://app.trading.vukem.dev", true},
		{"trailing slash is trimmed", true, "https://app.trading.vukem.dev/", true},
		{"plain http", true, "http://app.trading.vukem.dev", false},
		{"empty", true, "", false},
		{"localhost", true, "https://localhost", false},
		{"loopback address", true, "https://127.0.0.1:3000", false},
		{"path", true, "https://app.trading.vukem.dev/app", false},
		{"query", true, "https://app.trading.vukem.dev?x=1", false},
		{"development accepts anything", false, "http://localhost:3000", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateAppURL(tc.production, tc.url); (err == nil) != tc.ok {
				t.Fatalf("ValidateAppURL(%v, %q) = %v, want ok=%v", tc.production, tc.url, err, tc.ok)
			}
		})
	}
}
