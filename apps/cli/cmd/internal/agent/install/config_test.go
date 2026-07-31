package install

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNormalizeServerURL(t *testing.T) {

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "https becomes wss and gains the endpoint",
			input: "https://platform.example.com",
			want:  "wss://platform.example.com/agents",
		},
		{
			name:  "http becomes ws",
			input: "http://10.0.0.5:2002",
			want:  "ws://10.0.0.5:2002/agents",
		},
		{
			// The case named in the spec: no "//agents" may ever be produced.
			name:  "trailing slash does not double up",
			input: "https://platform.example.com/",
			want:  "wss://platform.example.com/agents",
		},
		{
			name:  "many trailing slashes",
			input: "https://platform.example.com///",
			want:  "wss://platform.example.com/agents",
		},
		{
			name:  "an endpoint already typed is not repeated",
			input: "https://platform.example.com/agents",
			want:  "wss://platform.example.com/agents",
		},
		{
			name:  "an endpoint with a trailing slash",
			input: "https://platform.example.com/agents/",
			want:  "wss://platform.example.com/agents",
		},
		{
			name:  "wss is preserved",
			input: "wss://platform.example.com/agents",
			want:  "wss://platform.example.com/agents",
		},
		{
			name:  "ws is preserved",
			input: "ws://localhost:3000",
			want:  "ws://localhost:3000/agents",
		},
		{
			// A bare host must not be read as scheme "platform.example.com".
			name:  "a bare host defaults to the secure scheme",
			input: "platform.example.com",
			want:  "wss://platform.example.com/agents",
		},
		{
			name:  "a bare host with a port",
			input: "10.0.0.5:2002",
			want:  "wss://10.0.0.5:2002/agents",
		},
		{
			name:  "a path prefix is kept",
			input: "https://example.com/docksight",
			want:  "wss://example.com/docksight/agents",
		},
		{
			name:  "a path prefix with a trailing slash",
			input: "https://example.com/docksight/",
			want:  "wss://example.com/docksight/agents",
		},
		{
			name:  "surrounding whitespace is ignored",
			input: "  https://platform.example.com  ",
			want:  "wss://platform.example.com/agents",
		},
	}

	for _, testCase := range cases {

		t.Run(testCase.name, func(t *testing.T) {

			got, err := NormalizeServerURL(testCase.input)

			if err != nil {
				t.Fatal(err)
			}

			if got != testCase.want {
				t.Fatalf("got %q, want %q", got, testCase.want)
			}

			_, afterScheme, _ := strings.Cut(got, "://")

			if strings.Contains(afterScheme, "//") {
				t.Fatalf("produced a doubled slash: %q", got)
			}
		})
	}
}

func TestNormalizeServerURLRejectsBadInput(t *testing.T) {

	cases := map[string]string{
		"empty":               "",
		"whitespace only":     "   ",
		"unsupported scheme":  "ftp://platform.example.com",
		"scheme with no host": "https://",
	}

	for name, input := range cases {

		t.Run(name, func(t *testing.T) {

			if _, err := NormalizeServerURL(input); err == nil {
				t.Fatalf("%q was accepted", input)
			}
		})
	}
}

func TestRenderConfigMatchesAgentSchema(t *testing.T) {

	layout := DefaultLayout()

	rendered := RenderConfig(layout, "wss://platform.example.com/agents")

	// The keys the agent's config loader unmarshals.
	for _, key := range []string{
		"agent:", "data_dir:", "identity_file:",
		"server:", "url:", "docker:", "socket:", "logging:", "level:",
	} {

		if !strings.Contains(rendered, key) {
			t.Errorf("rendered config is missing %q:\n%s", key, rendered)
		}
	}

	if !strings.Contains(rendered, `"wss://platform.example.com/agents"`) {
		t.Errorf("server url not quoted into the config:\n%s", rendered)
	}

	if !strings.Contains(rendered, `"/etc/docksight-agent/identity.json"`) {
		t.Errorf("identity path not written:\n%s", rendered)
	}
}

func TestWriteConfig(t *testing.T) {

	layout := DefaultLayout()
	layout.ConfigDir = filepath.Join(t.TempDir(), "docksight-agent")

	if err := WriteConfig(layout, "wss://platform.example.com/agents"); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(layout.ConfigPath())

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "wss://platform.example.com/agents") {
		t.Fatalf("config does not carry the platform URL:\n%s", string(content))
	}
}

// Reinstalling against a new platform must take effect.
func TestWriteConfigOverwrites(t *testing.T) {

	layout := DefaultLayout()
	layout.ConfigDir = t.TempDir()

	if err := WriteConfig(layout, "wss://old.example.com/agents"); err != nil {
		t.Fatal(err)
	}

	if err := WriteConfig(layout, "wss://new.example.com/agents"); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(layout.ConfigPath())

	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(content), "old.example.com") {
		t.Fatal("the previous platform URL survived")
	}
}

func TestConfigFileIsNotWorldReadable(t *testing.T) {

	if runtime.GOOS == "windows" {
		t.Skip("unix permission bits are not meaningful on windows")
	}

	layout := DefaultLayout()
	layout.ConfigDir = t.TempDir()

	if err := WriteConfig(layout, "wss://platform.example.com/agents"); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(layout.ConfigPath())

	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm()&0o004 != 0 {
		t.Fatalf("config is world readable: %04o", info.Mode().Perm())
	}
}

func TestIdentityExists(t *testing.T) {

	layout := DefaultLayout()
	layout.ConfigDir = t.TempDir()

	if IdentityExists(layout) {
		t.Fatal("a fresh host reported an existing identity")
	}

	if err := os.WriteFile(layout.IdentityPath(), []byte(`{"id":"abc"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if !IdentityExists(layout) {
		t.Fatal("a registered host reported no identity")
	}
}

func TestLayoutPaths(t *testing.T) {

	layout := DefaultLayout()

	if layout.BinaryPath != "/usr/local/bin/docksight-agent" {
		t.Errorf("binary path is %q", layout.BinaryPath)
	}

	if filepath.ToSlash(layout.ConfigPath()) != "/etc/docksight-agent/config.yaml" {
		t.Errorf("config path is %q", layout.ConfigPath())
	}

	if filepath.ToSlash(layout.IdentityPath()) != "/etc/docksight-agent/identity.json" {
		t.Errorf("identity path is %q", layout.IdentityPath())
	}
}
