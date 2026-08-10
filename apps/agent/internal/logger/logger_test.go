package logger

import (
	"bytes"
	"strings"
	"testing"
)

// Printf carries the human startup summary. It used to call fmt.Printf, which
// writes to stdout whatever Setup was given — under a Windows service there is
// no console, so the whole summary went to a handle nobody could read.
func TestPrintfWritesToTheConfiguredSink(t *testing.T) {

	var sink bytes.Buffer

	Setup("info", &sink)

	Printf("DockSight Agent %s\n", "v0.1.0")

	if !strings.Contains(sink.String(), "DockSight Agent v0.1.0") {
		t.Fatalf("Printf bypassed the configured sink; sink holds %q", sink.String())
	}
}

// The structured logger and Printf must land in the same place, or a service
// install would have its summary and its log in two different files.
func TestStructuredAndPlainOutputShareTheSink(t *testing.T) {

	var sink bytes.Buffer

	Setup("info", &sink)

	Printf("summary line\n")
	Info("structured line", "key", "value")

	content := sink.String()

	if !strings.Contains(content, "summary line") {
		t.Errorf("plain output missing: %q", content)
	}

	if !strings.Contains(content, "structured line") {
		t.Errorf("structured output missing: %q", content)
	}
}

// Warn must be filtered out at error level and kept at warn level: the
// plaintext-connection warning depends on it.
func TestLevelFiltering(t *testing.T) {

	var quiet bytes.Buffer

	Setup("error", &quiet)
	Warn("plaintext connection")

	if strings.Contains(quiet.String(), "plaintext connection") {
		t.Error("warn survived an error-level filter")
	}

	var loud bytes.Buffer

	Setup("warn", &loud)
	Warn("plaintext connection")

	if !strings.Contains(loud.String(), "plaintext connection") {
		t.Error("warn was dropped at warn level")
	}
}

// A nil sink must fall back to stdout rather than panicking on first write.
func TestSetupWithNilSinkDoesNotPanic(t *testing.T) {

	Setup("info", nil)

	Printf("")
	Info("still working")
}
