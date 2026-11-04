package trading

import (
	"os"
	"testing"
)

const wireFile = "../../resources/js/gen/wire.ts"

// TestWireTypesAreCurrent fails when a Go wire struct changed without regenerating the TypeScript.
// Regenerate with: go test ./internal/trading -run Wire -update.
func TestWireTypesAreCurrent(t *testing.T) {
	got := WireTypeScript()
	if *update {
		if err := os.MkdirAll("../../resources/js/gen", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(wireFile, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(wireFile)
	if err != nil {
		t.Fatalf("missing %s (run with -update): %v", wireFile, err)
	}
	if string(want) != got {
		t.Fatal("resources/js/gen/wire.ts is out of date: run go test ./internal/trading -run Wire -update")
	}
}
