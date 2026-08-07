package ui

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func TestASCIIFallbackSymbols(t *testing.T) {
	SetUnicode(false)
	out := captureStdout(t, func() {
		Success("ok")
		Error("bad")
		Info("info")
		Warning("warn")
	})
	for _, want := range []string{"[OK] ok", "[!!] bad", "-> info", "[!] warn"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Fatalf("expected %q in output %q", want, out)
		}
	}
}

func TestUnicodeSymbols(t *testing.T) {
	SetUnicode(true)
	out := captureStdout(t, func() {
		Success("ok")
		Error("bad")
		Info("info")
		Warning("warn")
	})
	for _, want := range []string{"✓ ok", "✗ bad", "→ info", "⚠ warn"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Fatalf("expected %q in output %q", want, out)
		}
	}
}
