package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRotatingFileCreatesMissingDirectory(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "DockSight", "logs", "agent.log")

	file, err := NewRotatingFile(path)

	if err != nil {
		t.Fatal(err)
	}

	defer file.Close()

	if _, err := file.Write([]byte("hello\n")); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("log file was not created: %v", err)
	}

	if string(content) != "hello\n" {
		t.Fatalf("got %q", content)
	}
}

// An agent runs unattended for months. Without a size bound, its log is a slow
// disk-exhaustion bug on the host DockSight is meant to be monitoring.
func TestRotationBoundsDiskUse(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "agent.log")

	file, err := NewRotatingFile(path)

	if err != nil {
		t.Fatal(err)
	}

	defer file.Close()

	// Shrink the limits so the test does not have to write 10 MiB.
	file.maxBytes = 64
	file.keep = 3

	line := strings.Repeat("x", 32) + "\n"

	for index := 0; index < 40; index++ {
		if _, err := file.Write([]byte(line)); err != nil {
			t.Fatalf("write %d: %v", index, err)
		}
	}

	entries, err := os.ReadDir(dir)

	if err != nil {
		t.Fatal(err)
	}

	// The current file plus at most `keep` rotated ones.
	if len(entries) > file.keep+1 {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Fatalf("%d files retained (%v), want at most %d", len(entries), names, file.keep+1)
	}

	// The oldest generation must have been discarded, not kept forever.
	if _, err := os.Stat(fmt.Sprintf("%s.%d", path, file.keep+1)); !os.IsNotExist(err) {
		t.Errorf("generation %d was not discarded", file.keep+1)
	}
}

// Rotation must not lose the line that triggered it.
func TestRotationPreservesTheTriggeringWrite(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "agent.log")

	file, err := NewRotatingFile(path)

	if err != nil {
		t.Fatal(err)
	}

	defer file.Close()

	file.maxBytes = 16

	if _, err := file.Write([]byte("first entry that fills the file\n")); err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write([]byte("second entry\n")); err != nil {
		t.Fatal(err)
	}

	current, err := os.ReadFile(path)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(current), "second entry") {
		t.Fatalf("the write that triggered rotation was lost; file holds %q", current)
	}

	rotated, err := os.ReadFile(path + ".1")

	if err != nil {
		t.Fatalf("rotated file missing: %v", err)
	}

	if !strings.Contains(string(rotated), "first entry") {
		t.Errorf("rotated file holds %q", rotated)
	}
}

// Reopening must continue an existing file rather than truncate it: a service
// restart should not erase the log that explains why it restarted.
func TestReopenAppendsRatherThanTruncates(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "agent.log")

	first, err := NewRotatingFile(path)

	if err != nil {
		t.Fatal(err)
	}

	if _, err := first.Write([]byte("before restart\n")); err != nil {
		t.Fatal(err)
	}

	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := NewRotatingFile(path)

	if err != nil {
		t.Fatal(err)
	}

	defer second.Close()

	if _, err := second.Write([]byte("after restart\n")); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "before restart") {
		t.Fatalf("reopening truncated the log; file holds %q", content)
	}

	if !strings.Contains(string(content), "after restart") {
		t.Fatalf("file holds %q", content)
	}
}
