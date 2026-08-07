package system

import "testing"

func TestPrimaryIPv4ReturnsSomething(t *testing.T) {
	got := PrimaryIPv4()
	if got == "" {
		t.Fatal("PrimaryIPv4 returned empty string")
	}
	// Either a dotted IPv4 or the localhost fallback.
	if got != "localhost" && !looksLikeIPv4(got) {
		t.Fatalf("unexpected address %q", got)
	}
}

func looksLikeIPv4(s string) bool {
	// Minimal shape check: four dotted decimal groups.
	dots := 0
	for _, c := range s {
		if c == '.' {
			dots++
		}
	}
	return dots == 3
}
