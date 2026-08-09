package install

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

// Layout describes where the agent lives on a host.
type Layout struct {
	// BinaryPath is the installed agent executable.
	BinaryPath string

	// ConfigDir holds config.yaml and identity.json.
	ConfigDir string

	// ServiceName is the systemd unit.
	ServiceName string

	// DockerSocket is the Engine socket the agent talks to.
	DockerSocket string

	// LogLevel is the agent's log verbosity.
	LogLevel string
}

// DefaultLayout is the standard agent installation.
func DefaultLayout() Layout {

	return Layout{
		BinaryPath:   "/usr/local/bin/docksight-agent",
		ConfigDir:    "/etc/docksight-agent",
		ServiceName:  "docksight-agent.service",
		DockerSocket: "/var/run/docker.sock",
		LogLevel:     "info",
	}
}

// ConfigPath is the agent configuration file.
//
// These paths are joined with path, not filepath: they are written into a
// YAML file read by an agent running on Linux, and a CLI built for Windows
// would otherwise emit "\etc\docksight-agent\identity.json" into it.
func (l Layout) ConfigPath() string {
	return path.Join(l.ConfigDir, "config.yaml")
}

// IdentityPath is where the agent persists the identity the platform
// assigns it.
func (l Layout) IdentityPath() string {
	return path.Join(l.ConfigDir, "identity.json")
}

// agentEndpoint is the WebSocket path the platform exposes for agents.
const agentEndpoint = "agents"

// NormalizeServerURL turns whatever the operator typed into the WebSocket URL
// the agent connects to.
//
//	https://platform.example.com        -> wss://platform.example.com/agents
//	http://10.0.0.5:2002                -> ws://10.0.0.5:2002/agents
//	platform.example.com                -> wss://platform.example.com/agents
//	https://platform.example.com/       -> wss://platform.example.com/agents
//	wss://platform.example.com/agents   -> unchanged
//
// Slashes are joined with path.Join, so a trailing slash can never produce
// the "//agents" path that a naive concatenation would.
func NormalizeServerURL(raw string) (string, error) {

	trimmed := strings.TrimSpace(raw)

	if trimmed == "" {
		return "", fmt.Errorf("the DockSight platform URL is required")
	}

	// A bare host has no scheme, and url.Parse would read "host:2002" as
	// scheme "host". Default to the secure scheme.
	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}

	parsed, err := url.Parse(trimmed)

	if err != nil {
		return "", fmt.Errorf("%q is not a valid URL: %w", raw, err)
	}

	scheme, err := websocketScheme(parsed.Scheme)

	if err != nil {
		return "", err
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("%q has no host", raw)
	}

	cleaned := path.Clean("/" + strings.Trim(parsed.Path, "/"))

	if cleaned == "/" {
		cleaned = ""
	}

	// Only append the endpoint when the operator has not already typed it.
	if !strings.HasSuffix(cleaned, "/"+agentEndpoint) {
		cleaned = path.Join(cleaned, agentEndpoint)
	}

	normalized := url.URL{
		Scheme: scheme,
		Host:   parsed.Host,
		Path:   "/" + strings.TrimPrefix(cleaned, "/"),
	}

	return normalized.String(), nil
}

// websocketScheme maps an entered scheme onto its WebSocket equivalent.
func websocketScheme(scheme string) (string, error) {

	switch strings.ToLower(scheme) {

	case "https", "wss":
		return "wss", nil

	case "http", "ws":
		return "ws", nil

	default:
		return "", fmt.Errorf(
			"unsupported scheme %q, use http, https, ws or wss",
			scheme,
		)
	}
}

// configTemplate matches the schema the agent parses in
// apps/agent/internal/config. Values are quoted so a URL or path containing
// YAML syntax cannot corrupt the file.
const configTemplate = `# DockSight Agent configuration
# Written by "docksight agent install". Edit and restart the service:
#   sudo systemctl restart %s

agent:
  data_dir: %q
  identity_file: %q

server:
  # DockSight platform WebSocket endpoint for agent registration
  url: %q

docker:
  socket: %q

logging:
  level: %q
`

// RenderConfig builds the agent configuration file.
func RenderConfig(layout Layout, serverURL string) string {

	return fmt.Sprintf(
		configTemplate,
		layout.ServiceName,
		layout.ConfigDir,
		layout.IdentityPath(),
		serverURL,
		layout.DockerSocket,
		layout.LogLevel,
	)
}

// WriteConfig creates the configuration directory and writes config.yaml.
//
// The file is rewritten on every install so a changed platform URL takes
// effect, but identity.json is never touched: it carries the identity the
// platform issued, and replacing it would register the host a second time and
// orphan its history.
func WriteConfig(layout Layout, serverURL string) error {

	if err := os.MkdirAll(layout.ConfigDir, 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", layout.ConfigDir, err)
	}

	content := RenderConfig(layout, serverURL)

	// 0640: the file names the platform the host reports to, and later gains
	// enrollment credentials. Readable by root, not by the world.
	if err := os.WriteFile(layout.ConfigPath(), []byte(content), 0o640); err != nil {
		return fmt.Errorf("failed to write %s: %w", layout.ConfigPath(), err)
	}

	return nil
}

// ReadServerURL returns the platform URL recorded in an existing
// installation, so an update does not require the operator to retype it.
//
// This reads the one key it needs rather than unmarshalling YAML: the CLI has
// no YAML dependency, and a config the agent accepts but this cannot parse
// would block updates for no good reason.
func ReadServerURL(layout Layout) (string, error) {

	file, err := os.Open(layout.ConfigPath())

	if err != nil {
		return "", err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	inServer := false

	for scanner.Scan() {

		line := scanner.Text()

		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// A key at column zero starts a new top-level section.
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inServer = strings.HasPrefix(trimmed, "server:")
			continue
		}

		if !inServer {
			continue
		}

		value, found := strings.CutPrefix(trimmed, "url:")

		if !found {
			continue
		}

		return strings.Trim(strings.TrimSpace(value), `"'`), nil
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("no server.url found in %s", layout.ConfigPath())
}

// UpdateServerURL changes only server.url in an existing config file. Other
// operator-managed settings are preserved, and identity.json is never opened.
func UpdateServerURL(layout Layout, serverURL string) error {

	content, err := os.ReadFile(layout.ConfigPath())

	if err != nil {
		return err
	}

	info, err := os.Stat(layout.ConfigPath())

	if err != nil {
		return err
	}

	updated, err := replaceServerURL(string(content), serverURL)

	if err != nil {
		return err
	}

	staged, err := os.CreateTemp(layout.ConfigDir, ".config-")

	if err != nil {
		return err
	}

	stagedPath := staged.Name()

	defer os.Remove(stagedPath)

	if _, err := staged.WriteString(updated); err != nil {
		staged.Close()
		return err
	}

	if err := staged.Close(); err != nil {
		return err
	}

	if err := os.Chmod(stagedPath, info.Mode().Perm()); err != nil {
		return err
	}

	if err := os.Rename(stagedPath, layout.ConfigPath()); err != nil {
		return fmt.Errorf("failed to update %s: %w", layout.ConfigPath(), err)
	}

	return nil
}

func replaceServerURL(content string, serverURL string) (string, error) {

	lines := strings.SplitAfter(content, "\n")
	inServer := false

	for index, rawLine := range lines {

		line := rawLine
		ending := ""

		switch {
		case strings.HasSuffix(line, "\r\n"):
			line = strings.TrimSuffix(line, "\r\n")
			ending = "\r\n"
		case strings.HasSuffix(line, "\n"):
			line = strings.TrimSuffix(line, "\n")
			ending = "\n"
		}

		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inServer = strings.HasPrefix(trimmed, "server:")
			continue
		}

		if !inServer {
			continue
		}

		if _, found := strings.CutPrefix(trimmed, "url:"); !found {
			continue
		}

		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		lines[index] = indent + "url: " + strconv.Quote(serverURL) + ending

		return strings.Join(lines, ""), nil
	}

	return "", fmt.Errorf("no server.url found in config")
}

// IdentityExists reports whether this host has already registered.
func IdentityExists(layout Layout) bool {

	info, err := os.Stat(layout.IdentityPath())

	return err == nil && info.Size() > 0
}
