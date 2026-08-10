package scm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTail(t *testing.T) {

	cases := []struct {
		name  string
		text  string
		lines int
		want  string
	}{
		{
			name:  "fewer lines than asked for",
			text:  "one\ntwo\n",
			lines: 40,
			want:  "one\ntwo",
		},
		{
			name:  "the last lines only",
			text:  "one\ntwo\nthree\nfour\n",
			lines: 2,
			want:  "three\nfour",
		},
		{
			name:  "an empty file",
			text:  "",
			lines: 40,
			want:  "",
		},
		{
			name:  "trailing newlines do not become a blank line",
			text:  "one\n\n",
			lines: 1,
			want:  "one",
		},
	}

	for _, testCase := range cases {

		t.Run(testCase.name, func(t *testing.T) {

			if got := tail(testCase.text, testCase.lines); got != testCase.want {
				t.Fatalf("got %q, want %q", got, testCase.want)
			}
		})
	}
}

// Verification greps the agent's own log, so the tail must carry the marker
// lines it looks for rather than truncating them off the end.
func TestTailFileKeepsConnectionMarkers(t *testing.T) {

	path := filepath.Join(t.TempDir(), "agent.log")

	var builder strings.Builder

	for range 100 {
		builder.WriteString("level=INFO msg=\"connecting to docksight server\"\n")
	}

	builder.WriteString("level=INFO msg=\"websocket connected\"\n")

	if err := os.WriteFile(path, []byte(builder.String()), 0o640); err != nil {
		t.Fatal(err)
	}

	content, err := tailFile(path, 40)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(content, "websocket connected") {
		t.Fatalf("the tail lost the marker:\n%s", content)
	}

	if lines := strings.Count(content, "\n") + 1; lines != 40 {
		t.Fatalf("read %d lines, want 40", lines)
	}
}

func TestTailFileMissing(t *testing.T) {

	if _, err := tailFile(filepath.Join(t.TempDir(), "absent.log"), 40); err == nil {
		t.Fatal("a missing log file was read successfully")
	}
}

func TestDefaultLogPathFollowsProgramData(t *testing.T) {

	t.Setenv("ProgramData", `C:\ProgramData`)

	// The agent writes to this exact path — see ServiceLogPath in
	// apps/agent/internal/logger/sink_windows.go. If the two drift,
	// verification reads an empty file and never confirms a connection.
	want := filepath.Join(`C:\ProgramData`, "DockSight", "logs", "agent.log")

	if got := DefaultLogPath(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
