package cmd

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Open-Source-Kigali/docksight/apps/cli/cmd/internal/agent/install"
)

type fakeAgentRestarter struct {
	calls int
	unit  string
}

func (f *fakeAgentRestarter) Restart(_ context.Context, unit string) error {
	f.calls++
	f.unit = unit
	return nil
}

func TestReadAgentConfig(t *testing.T) {

	layout := install.DefaultLayout()
	layout.ConfigDir = t.TempDir()

	if err := install.WriteConfig(layout, "wss://platform.example.com/agents"); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(layout.IdentityPath(), []byte(`{"id":"registered"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	details, err := readAgentConfig(layout)

	if err != nil {
		t.Fatal(err)
	}

	if details.platformURL != "wss://platform.example.com/agents" {
		t.Errorf("platform URL = %q", details.platformURL)
	}
	if details.binaryPath != layout.BinaryPath {
		t.Errorf("binary path = %q", details.binaryPath)
	}
	if details.configPath != layout.ConfigPath() {
		t.Errorf("config path = %q", details.configPath)
	}
	if details.unitName != layout.ServiceName {
		t.Errorf("unit name = %q", details.unitName)
	}
	if details.dockerSocket != layout.DockerSocket {
		t.Errorf("Docker socket = %q", details.dockerSocket)
	}
	if !details.registered {
		t.Error("registered agent reported no")
	}
}

func TestSetAgentServerURLNormalizesAndRestarts(t *testing.T) {

	layout := install.DefaultLayout()
	layout.ConfigDir = t.TempDir()

	if err := install.WriteConfig(layout, "wss://old.example.com/agents"); err != nil {
		t.Fatal(err)
	}

	restarter := &fakeAgentRestarter{}

	got, err := setAgentServerURL(
		context.Background(),
		layout,
		restarter,
		"new.example.com",
	)

	if err != nil {
		t.Fatal(err)
	}

	if got != "wss://new.example.com/agents" {
		t.Fatalf("normalized URL = %q", got)
	}

	stored, err := install.ReadServerURL(layout)

	if err != nil {
		t.Fatal(err)
	}

	if stored != got {
		t.Fatalf("stored URL = %q, want %q", stored, got)
	}
	if restarter.calls != 1 || restarter.unit != layout.ServiceName {
		t.Fatalf("restart calls = %d, unit = %q", restarter.calls, restarter.unit)
	}
}

func TestSetAgentServerURLRejectsInvalidInputWithoutWriting(t *testing.T) {

	layout := install.DefaultLayout()
	layout.ConfigDir = t.TempDir()

	if err := install.WriteConfig(layout, "wss://old.example.com/agents"); err != nil {
		t.Fatal(err)
	}

	before, err := os.ReadFile(layout.ConfigPath())

	if err != nil {
		t.Fatal(err)
	}

	for _, invalid := range []string{"", "not a url"} {

		restarter := &fakeAgentRestarter{}

		if _, err := setAgentServerURL(
			context.Background(),
			layout,
			restarter,
			invalid,
		); err == nil {
			t.Fatalf("invalid URL %q was accepted", invalid)
		}

		if restarter.calls != 0 {
			t.Fatalf("service restarted %d times after invalid input %q", restarter.calls, invalid)
		}
	}

	after, err := os.ReadFile(layout.ConfigPath())

	if err != nil {
		t.Fatal(err)
	}

	if string(after) != string(before) {
		t.Fatalf("config changed after invalid input:\n%s", after)
	}
	if !strings.Contains(string(after), "old.example.com") {
		t.Fatal("original URL was not preserved")
	}
}
