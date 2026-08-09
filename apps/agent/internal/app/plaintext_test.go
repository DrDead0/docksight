package app

import "testing"

func TestIsLoopbackHost(t *testing.T) {
	t.Parallel()

	cases := []struct {
		host string
		want bool
	}{
		{"localhost", true},
		{"LOCALHOST", true},
		{"127.0.0.1", true},
		{"127.0.0.2", true},
		{"::1", true},
		{"0:0:0:0:0:0:0:1", true},
		{"localhost.evil.com", false},
		{"notlocalhost", false},
		{"127.0.0.1.nip.io", false},
		{"example.com", false},
		{"192.168.1.1", false},
		{"10.0.0.1", false},
		{"", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.host, func(t *testing.T) {
			t.Parallel()
			got := isLoopbackHost(tc.host)
			if got != tc.want {
				t.Fatalf("isLoopbackHost(%q) = %v, want %v", tc.host, got, tc.want)
			}
		})
	}
}

func TestWarnIfPlaintextServerURL_schemeAndHost(t *testing.T) {
	t.Parallel()

	// Smoke: pure functions only — warn path must not panic on common URLs.
	urls := []string{
		"ws://localhost:3000/agents",
		"ws://127.0.0.1:3000/agents",
		"ws://[::1]:3000/agents",
		"ws://localhost.evil.com/agents",
		"ws://agent.example.com/agents",
		"wss://agent.example.com/agents",
		"https://not-a-ws.example",
		"://bad",
		"",
	}
	for _, raw := range urls {
		warnIfPlaintextServerURL(raw)
	}
}
